import os
from dotenv import load_dotenv
from flask_sqlalchemy import SQLAlchemy
from sqlalchemy import text
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.orm import relationship



# 加载环境变量
load_dotenv()

# 初始化SQLAlchemy实例
db = SQLAlchemy()

class Role(db.Model):
    __tablename__ = 'roles'
    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    name = db.Column(db.String(50), nullable=False, comment='角色名称')
    alias = db.Column(db.String(50), nullable=False, unique=True, comment='角色别名，用于关联users表')
    created_at = db.Column(db.TIMESTAMP, nullable=False, default=db.func.current_timestamp(), comment='创建时间')
    status = db.Column(db.Boolean, nullable=False, server_default=db.text('1'), comment='状态，0-冻结，1-启用')
    remark = db.Column(db.Text, comment='备注')
    menus_ids = db.Column(db.String(255), default=None, comment='菜单ID，多个以逗号分隔')
    menus_tree = db.Column(db.String(255), default=None, comment='菜单ID，多个以逗号分隔')
    

class User(db.Model):
    __tablename__ = 'users'
    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    nickname = db.Column(db.String(10), nullable=False, comment='用户昵称，支持10位中英文')
    account = db.Column(db.String(12), nullable=False, unique=True, comment='账号，8到12位数字和字母')
    avatar_url = db.Column(db.String(255), default=None, comment='头像URL，图片小于10K')
    phone_number = db.Column(db.String(15), default=None, comment='手机号')
    status = db.Column(db.Boolean, nullable=False, server_default=db.text('1'), comment='用户状态，0-冻结，1-启用')
    created_at = db.Column(db.TIMESTAMP, nullable=False, default=db.func.current_timestamp(), comment='创建时间')
    password = db.Column(db.String(255), nullable=False, comment='密码')
    init_pwd_status = db.Column(db.Boolean, nullable=False, server_default=db.text('1'))
    

class Menu(db.Model):
    __tablename__ = 'menus'
    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    parent_id = db.Column(db.Integer, db.ForeignKey('menus.id'), nullable=True)
    name = db.Column(db.String(50), default=None)
    title = db.Column(db.String(255), default=None)
    path = db.Column(db.String(255), default=None)
    component = db.Column(db.String(255), default=None)
    icon = db.Column(db.String(50), default=None)
    order_num = db.Column(db.Integer, nullable=False, default=0)
    status = db.Column(db.Enum('ACTIVE', 'INACTIVE'), nullable=False, server_default='ACTIVE')
    created_at = db.Column(db.TIMESTAMP, nullable=False, default=db.func.current_timestamp())
    updated_at = db.Column(db.TIMESTAMP, nullable=False, default=db.func.current_timestamp(), onupdate=db.func.current_timestamp())

    parent = relationship("Menu", remote_side=[id], back_populates="children")
    children = relationship("Menu", cascade="all, delete-orphan", back_populates="parent")


    def delete(self, session):
        """
        删除当前菜单及其所有子菜单。
        参数:
            session (Session): SQLAlchemy会话对象，用于执行数据库操作。
        """
        # 首先，递归删除所有子菜单
        for child in self.children:
            child.delete(session)
        
        # 然后，删除当前菜单
        session.delete(self)
        session.commit()


class UserRole(db.Model):
    __tablename__ = 'user_role'
    user_id = db.Column(db.Integer, db.ForeignKey('users.id'), primary_key=True)
    role_id = db.Column(db.Integer, db.ForeignKey('roles.id'), primary_key=True)

    # 添加唯一性约束
    __table_args__ = (db.UniqueConstraint('role_id', 'user_id', name='_role_user_uc'),)



class RoleMenu(db.Model):
    __tablename__ = 'role_menu'
    role_id = db.Column(db.Integer, db.ForeignKey('roles.id'), primary_key=True)
    menu_id = db.Column(db.String(255), db.ForeignKey('menus.id'), primary_key=True)

    # 添加唯一性约束
    __table_args__ = (db.UniqueConstraint('role_id', 'menu_id', name='_role_menu_uc'),)




class DialogLogs(db.Model):
    __tablename__ = 'dialog_logs'

    id = db.Column(db.Integer, primary_key=True)
    user_id = db.Column(db.Integer, db.ForeignKey('users.id'), nullable=False)
    directory_id = db.Column(db.Integer, db.ForeignKey('dialog_directories_user.id'), nullable=False)
    input_content = db.Column(db.Text, nullable=False)
    output_content = db.Column(db.Text, nullable=False)
    timestamp = db.Column(db.TIMESTAMP, nullable=False, default=db.func.current_timestamp())
    status_code = db.Column(db.Integer)
    status_message = db.Column(db.String(255))
    err_code = db.Column(db.Integer)
    err_msg = db.Column(db.String(255))
    labels = db.Column(db.Text)
    type = db.Column(db.Integer)




class DialogDirectoriesUser(db.Model):
    __tablename__ = 'dialog_directories_user'

    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    user_id = db.Column(db.Integer, db.ForeignKey('users.id', ondelete='CASCADE'), nullable=False)
    directory_name = db.Column(db.String(255), nullable=False)
    created_at = db.Column(db.TIMESTAMP, nullable=False, server_default=db.func.current_timestamp())

    # This will create a backref 'directories' on the User model if it's defined correctly.
    # The relationship is not strictly necessary for the model to work but can be useful for querying.



class ModelConfig(db.Model):
    __tablename__ = 'model_configs'
    
    id = db.Column(db.Integer, primary_key=True)
    user_id = db.Column(db.Integer, db.ForeignKey('users.id'), nullable=False)
    cdK = db.Column(db.String(255), nullable=False)
    use_stream = db.Column(db.Boolean, default=False)
    content_type = db.Column(db.Integer, default=1)
    class_id = db.Column(db.String(10))
    proxy_url = db.Column(db.String(255))
    bypass_uri = db.Column(db.Text)
    config_network_status = db.Column(db.Integer, default=0)
    local_log_path = db.Column(db.String(255))
    remote_log_path = db.Column(db.String(255))
    api_log_status = db.Column(db.Integer, default=0)
    expiry_date = db.Column(db.DateTime)
    created_at = db.Column(db.TIMESTAMP, nullable=False, default=db.func.current_timestamp())
    updated_at = db.Column(db.TIMESTAMP, nullable=False, default=db.func.current_timestamp(), onupdate=db.func.current_timestamp())



# 图片模型
class Image(db.Model):
    __tablename__ = 'images'
    
    id = db.Column(db.Integer, primary_key=True)
    user_id = db.Column(db.Integer, db.ForeignKey('users.id'), nullable=False)
    filename = db.Column(db.String(255), nullable=False)
    path = db.Column(db.String(255), nullable=False)
    status = db.Column(db.Boolean, nullable=False, server_default=db.text('1'))
    type = db.Column(db.String(10))
    created_at = db.Column(db.TIMESTAMP, nullable=False, server_default=db.func.current_timestamp())



class ContactUs(db.Model):
    __tablename__ = 'contact_us'
    
    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    contact_info = db.Column(db.Text)
    image_path = db.Column(db.String(255))
    created_at = db.Column(db.TIMESTAMP, nullable=False, default=db.func.current_timestamp())
    updated_at = db.Column(db.TIMESTAMP, nullable=False, default=db.func.current_timestamp(), onupdate=db.func.current_timestamp())


class AboutUs(db.Model):
    __tablename__ = 'about_us'
    
    id = db.Column(db.Integer, primary_key=True, autoincrement=True)
    title = db.Column(db.Text)
    content = db.Column(db.Text)
    created_at = db.Column(db.TIMESTAMP, nullable=False, default=db.func.current_timestamp())
    updated_at = db.Column(db.TIMESTAMP, nullable=False, default=db.func.current_timestamp(), onupdate=db.func.current_timestamp())




def init_db(app):
    """初始化数据库连接与配置"""
    app.config['SQLALCHEMY_DATABASE_URI'] = os.getenv('SQLALCHEMY_DATABASE_URI')
    app.config['SQLALCHEMY_POOL_SIZE'] = int(os.getenv('SQLALCHEMY_POOL_SIZE', 10))
    app.config['SQLALCHEMY_MAX_OVERFLOW'] = int(os.getenv('SQLALCHEMY_MAX_OVERFLOW', 20))
    app.config['SQLALCHEMY_POOL_RECYCLE'] = int(os.getenv('SQLALCHEMY_POOL_RECYCLE', 3600))
    app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False

    db.init_app(app)

    # 确保每次请求结束时清理数据库会话
    @app.teardown_appcontext
    def shutdown_session(exception=None):
        db.session.remove()

def execute_query(query, params=None):
    """安全执行SQL查询"""
    with db.engine.begin() as connection:
        if params:
            result = connection.execute(text(query), params)
        else:
            result = connection.execute(text(query))
        try:
            return result.fetchall()
        except SQLAlchemyError as e:
            db.session.rollback()
            raise e

def get_record_by_field(model, field_name, value):
    '''
    根据字段名称和值从数据库中获取User模型的实例。
    
    :param field_name: 字段名称，如'email'或'username'
    :param value: 字段对应的值
    :return: User实例或None
    '''
    if hasattr(model, field_name):  # 确保User模型中有该字段
        column = getattr(model, field_name)
        print(column)
        return model.query.filter(column == value).first()  # 使用query和filter方法查询
    else:
        raise AttributeError(f"Field '{field_name}' does not exist on model '{model}'.")



def get_record_by_field_all(model, field_name, value):
    '''
    根据字段名称和值从数据库中获取User模型的实例。
    
    :param field_name: 字段名称，如'email'或'username'
    :param value: 字段对应的值
    :return: User实例或None
    '''
    if hasattr(model, field_name):  # 确保User模型中有该字段
        column = getattr(model, field_name)
        print(column)
        return model.query.filter(column == value).all()  # 使用query和filter方法查询
    else:
        raise AttributeError(f"Field '{field_name}' does not exist on model '{model}'.")