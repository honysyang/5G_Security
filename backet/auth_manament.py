import os
from functools import wraps
from flask import request, jsonify, current_app,Flask, request, jsonify, send_file
from flask_jwt_extended import (
    JWTManager, jwt_required, create_access_token, get_jwt_identity
)
from werkzeug.security import generate_password_hash, check_password_hash
from database_management import User, Role, db, get_record_by_field, Menu, UserRole, RoleMenu, get_record_by_field_all
from validators import validate_password,validate_nickname,validate_account,validate_avatar_url,validate_phone_number
from captcha.image import ImageCaptcha
import random
import string
from werkzeug.utils import secure_filename
from flask import session

import secrets
import uuid
import base64



# 初始化JWT管理器
jwt = JWTManager()


# 验证码生成器初始化
image_captcha = ImageCaptcha()


def generate_secret_key(length=32):
    """Generate a secure random secret key."""
    alphabet = string.ascii_letters + string.digits + string.punctuation
    return ''.join(secrets.choice(alphabet) for _ in range(length))

# 生成一个32字符长度的随机密钥
secret_key = generate_secret_key()
print(f"Generated Secret Key: {secret_key}")



def init_jwt(app):
    """初始化JWT相关配置"""
    app.config['JWT_SECRET_KEY'] = os.getenv('JWT_SECRET_KEY')  # 应从环境变量中读取
    jwt.init_app(app)

def authenticate(account, password):
    """用户认证函数"""

    
    user = User.query.filter_by(account=account).first()
    
    print(user)
    if user and check_password_hash(user.password, password):
        return user
    return None


def register_routes(app):
    """注册认证相关的路由"""

    app.config['CAPTCHA_IMAGE_DIR'] = os.getenv('CAPTCHA_IMAGE_DIR')
    app.config['SECRET_KEY'] = secret_key 

    def generate_session_identifier():
    # 这里创建一个唯一的标识符并存储在session中
        if 'custom_session_id' not in session:
            session['custom_session_id'] = str(uuid.uuid4())  # 示例，实际可以根据需要生成
        return session['custom_session_id']



    def generate_captcha_text(length=4):
        """生成随机验证码文本"""
        return ''.join(random.choices(string.ascii_uppercase + string.digits, k=length))

    def generate_captcha_image(text):
        """生成验证码图片并保存，返回图片路径"""
        filename = secure_filename(text + '.png')
        filepath = os.path.join(app.config['CAPTCHA_IMAGE_DIR'], filename)
        image_captcha.write(text, filepath)
        return filepath

    @app.route('/ai/captcha', methods=['POST'])
    def get_captcha():
        try:
            """生成并返回验证码图片及会话标识"""
            text = generate_captcha_text()
            filepath = generate_captcha_image(text)
            
            # 将验证码答案存储在session中，并生成或重置一个session id
            session['captcha'] = text
            session.modified = True  # 标记session已修改
            session['id'] = generate_session_identifier()  # 获取或生成session id
            print(session['captcha'])  # 打印会话对象的所有属性和方法
            '''
            # 返回验证码图片及session id给客户端
            response = send_file(filepath, mimetype='image/png')

            response.set_cookie('session_id', session['id'])  # 设置cookie传递session id给客户端
            return response
            '''
            # 读取图片文件并转换为Base64编码
            with open(filepath, 'rb') as img_file:
                img_base64 = base64.b64encode(img_file.read()).decode('utf-8')

            # 构建JSON响应数据
            response_data = {
                    'code':200,
                    'message':"ok",
                    'data':{
                        'img_base64': img_base64,
                        'session_id': session['id']
                    }
                }

            # 清理本地文件系统中的图片文件
            os.remove(filepath)

            # 返回JSON响应
            return jsonify(response_data)
        except Exception as e:
            return jsonify({
                'code':200,
                "error": "An error occurred the captcha."
                }), 500




    #用户登录
    @app.route('/ai/login', methods=['POST'])
    def login():
        data = request.get_json()
        account = data.get('user')
        password = data.get('password')
        captcha_solution = data.get('captcha')  # 客户端提交的验证码答案
        #received_session_id = request.cookies.get('session_id')  # 从请求中获取session id

        
        
        try:
            
            '''
            # 确保session id有效且与当前活跃session匹配
            if received_session_id != session['id']:
                return jsonify({
                "status": "error",
                "code": 400,
                "message": "Invalid session_id"
                }), 400
            '''
            # 验证码校验
            if captcha_solution.lower() != session['captcha'].lower():  # 不区分大小写比较
                return jsonify({
                "status": "error",
                "code": 401,
                "message": "Incorrect captcha"
                }), 401

            
            user = authenticate(account, password)
            if user:
                access_token = create_access_token(identity=user.id)

                role_user = get_record_by_field(UserRole, 'user_id',user.id)
                #通过access_role获取到roles表中的name
                role = get_record_by_field(Role, 'id', role_user.role_id)
                
                if user.status == '0':
                    return jsonify({
                        "status": "error",
                        "code": 401,
                        "message": "Account is not active"
                    }), 401
                
                
                # 登录成功后清除session中的验证码信息
                session.pop('captcha', None)
                session.modified = True

                # 登录成功响应
                return jsonify({
                    "status": "success",
                    "code": 200,
                    "data": {
                        "access_token": access_token,
                        "role_id": role.id,
                        "user_id": user.id,
                        "init_pwd_status": user.init_pwd_status
                    }
                }), 200
                
            # 凭证无效响应
            return jsonify({
                "status": "error",
                "code": 401,
                "message": "Invalid credentials account"
            }), 401
        except Exception as e:
            # 异常处理
            print(e)
            return jsonify({
                "status": "error",
                "code": 500,
                "message": "Internal server error"
            }), 500


    #用户登出
    @app.route('/ai/logout', methods=['POST'])
    @jwt_required()
    def logout():
        jwt.unauthorize()
        return jsonify({
            "status": "success",
            "code": 200,
            "message": "Logged out"
        }), 200




    #用户登录输出菜单
    @app.route('/ai/me/menu', methods=['POST'])
    #@jwt_required()
    def menu():
        # 获取当前登录用户的id
        data = request.get_json()
        user_id = data.get('user_id')
        role_id = data.get('role_id')

        try:

            #通过role_id获得role表中的alias
            role = get_record_by_field(Role, 'id', role_id)
            if not role:
                return jsonify({"message": "Role not found"}), 404

            # 获取当前登录用户的权限
            role_name = role.alias

            '''
            #判断当前用户角色状态是否为1
            if role.status == 0:
                return jsonify({"message": "Role is not active"}), 404

            '''
           
            # 通过role_id获取角色关联的菜单ID列表
            role_menu_relations = get_record_by_field_all(RoleMenu, 'role_id', role_id)
            if not role_menu_relations:
                return jsonify({"message": "No menu permissions found for this role"}), 404

            print(role_menu_relations)
            print(type(role_menu_relations))
            # 从关联中提取menu_id
            menu_ids = [relation.menu_id for relation in role_menu_relations]  # 修改这里

            # 查询这些menu_id对应的菜单项
            menus = Menu.query.filter(Menu.id.in_(menu_ids)).all()


            print(menus)
            # 4. 构建菜单结构（这里简单示例，未处理层级关系）
            menu_list = []
            for menu in menus:
                menu_item = {
                    "id": menu.id,
                    "parent_Id": menu.parent_id,
                    "name": menu.name,
                    "path": menu.path,
                    "component": menu.component,
                    "meta": {  # 添加嵌套的"meta"字段
                        "icon": menu.icon,
                        "title": menu.title,
                        "role": role_name
                    }
                }
                menu_list.append(menu_item)

            return jsonify({
                "status": "success",
                "code": 200,
                "data": {
                    "menus": menu_list
                }
            })
        
        except Exception as e:
            return jsonify({
                "status": "error",
                "code": 500,
                "message": str(e)
            }), 500
        



    #展示用户的详细信息，，这里可以点击个人
    @app.route('/ai/me/display_info', methods=['POST'])
    #@jwt_required()
    def me():
        # 获取当前登录用户的id
        data = request.get_json()
        user_id = data.get('user_id')
        
        user = get_record_by_field(User, 'id', user_id)

        #获取其对应的role信息
        role_user = get_record_by_field(UserRole, 'user_id',user.id)
        
        role = get_record_by_field(Role, 'id', role_user.role_id)

        if user:
            return jsonify({
                "status": "success",
                "code": 200,
                "data": {
                    "username": user.nickname,
                    "role": role.alias,
                    "avatar_url": user.avatar_url if user.avatar_url else None,
                    "phone_number": user.phone_number if user.phone_number else None,
                    "status": user.status,
                    "create_time": user.create_time.strftime("%Y-%m-%d %H:%M:%S") if user.create_time else None,
                }
            }), 200
        return jsonify({
            "status": "error",
            "code": 404,
            "message": "User not found"
            }), 404



    #修改个人密码
    @app.route('/ai/me/change_password', methods=['PUT'])
    #@jwt_required()
    def change_password():
        data = request.get_json()
        user_id = data.get('user_id')
        '''
        current_user_id = get_jwt_identity()

        if current_user_id != user_id:
            return jsonify({
                "status": "error",
                "code": 401,
                "message": "Unauthorized"
            }), 401
        '''
        user = get_record_by_field(User, 'id',user_id)
        try:
            if user:
                data = request.get_json()
                old_password = data.get('old_password')
                new_password = data.get('new_password')
                two_password = data.get('two_password')
                if new_password != two_password:
                    return jsonify({
                        "status": "error",
                        "code": 400,
                        "message": "New passwords do not match"
                    }), 400
                if user and check_password_hash(user.password, old_password):
                    user.password = generate_password_hash(new_password)
                    user.init_pwd_status = 1
                    db.session.commit()
                    return jsonify({
                        "status": "success",
                        "code": 200,
                        "data": {
                            "message": "Password changed successfully"
                        }
                    }), 200
                return jsonify({
                    "status": "error",
                    "code": 400,
                    "message": "Old password is incorrect"
                }), 401
            return jsonify({
                "status": "error",
                "code": 404,
                "message": "User not found"
            }), 404
        except Exception as e:
            print(e)
            return jsonify({
                "status": "error",
                "code": 500,
                "message": "Internal server error"
            }), 500

    


    #修改用户个人的信息，这里的修改，只可以修改nickname、avatar_url、phone_number
    @app.route('/ai/me/change_info', methods=['PUT'])
    #@jwt_required()
    def change_info():
        data = request.get_json()
        user_id = data.get('user_id')
        '''
        current_user_id = get_jwt_identity()

        if current_user_id != user_id:
            return jsonify({
                "status": "error",
                "code": 401,
                "message": "Unauthorized"
            }), 401
        '''
        user = get_record_by_field(User, 'id',user_id)
        try:
            if user:
                data = request.get_json()
                avatar_url = data.get('avatar_url')
                phone_number = data.get('phone_number')
                nickname = data.get('nickname')
                if avatar_url:
                    if validate_avatar_url(avatar_url):
                        user.avatar_url = avatar_url
                    else:
                        return jsonify({
                            "status": "error",
                            "code": 400,
                            "message": "Invalid avatar URL"
                        }), 400
                if phone_number:
                    if validate_phone_number(phone_number):
                        user.phone_number = phone_number
                    else:
                        return jsonify({
                            "status": "error",
                            "code": 400,
                            "message": "Invalid phone number"
                        }), 400
                if nickname:
                    if validate_nickname(nickname):
                        user.nickname = nickname  
                    else:
                        return jsonify({
                            "status": "error",
                            "code": 400,
                            "message": "Invalid nickname"
                        }), 400
                db.session.commit()
                return jsonify({
                    "status": "success",
                    "code": 200,
                    "data": {
                        "message": "Info changed successfully"
                    }
                }), 200
            return jsonify({
                "status": "error",
                "code": 404,
                "message": "User not found"
            }), 404
        except Exception as e:
            print(e)
            return jsonify({
                "status": "error",
                "code": 500,
                "message": "Internal server error"
            }), 500



    


