from flask import Flask,jsonify
from flask_migrate import Migrate
from flask_jwt_extended import JWTManager
from dotenv import load_dotenv
import os
from database_management import init_db, db
from auth_manament import init_jwt, register_routes
from user_management import user_management_routes
from chat_management import chat_routes
from model_management import model_management_routes
from upload_management import upload_image_routes
from chat_test import chat_test_routes
from log_management import log_routes
from writing_management import writing_route
from flow_stat import flow_routes


# 加载环境变量
load_dotenv()

def create_app():
    app = Flask(__name__)
    
    # 配置Flask应用
    app.config['SQLALCHEMY_DATABASE_URI'] = os.getenv('SQLALCHEMY_DATABASE_URI')
    app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
    app.config['JWT_SECRET_KEY'] = os.getenv('JWT_SECRET_KEY')  # 从环境变量中读取JWT密钥

    # 初始化数据库
    init_db(app)
    
    #模型参数
    model_management_routes(app)

    # 初始化JWT
    init_jwt(app)

    log_routes(app)

    upload_image_routes(app)

    writing_route(app)
    
    #大模型聊天窗口
    chat_routes(app)

    chat_test_routes(app)

    #流量统计

    flow_routes(app)

    # 注册路由
    register_routes(app)

    # 用户和角色管理的路由
    user_management_routes(app)
    
    # 初始化数据库迁移工具
    Migrate(app, db)
    
    return app


app = create_app()


if __name__ == '__main__':
    # 启动Flask应用
    app.run(host='0.0.0.0', port=8001, debug=os.environ.get('FLASK_DEBUG', 'true').lower() == 'true')