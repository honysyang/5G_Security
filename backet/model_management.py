from auth_manament import init_jwt, register_routes,jwt_required
from flask import request, Response, jsonify
from database_management import db,DialogLogs,User,DialogDirectoriesUser,get_record_by_field,get_record_by_field_all,ModelConfig

class GlobalModel:

     def __init__(self, user_id):
        model_config = get_record_by_field(ModelConfig, 'user_id', user_id)
        if model_config:
            if model_config.use_stream == 1:
                self.chat_stream = True    # 标记聊天是否是流式
            else:
                self.chat_stream = False
            if model_config.config_network_status == 1:
                self.config_network_status = 1
                self.proxy_url = model_config.proxy_url
                self.bypass_uri = model_config.bypass_uri
            else:
                self.config_network_status = 0        # 标记是串联，还是并联
                self.proxy_url = model_config.proxy_url
                self.bypass_uri = model_config.bypass_uri

            if model_config.api_log_status == 1:
                self.api_log = model_config.remote_log_path       # 标记日志存储路径
            else:
                self.api_log = model_config.local_log_path       # 标记日志存储路径
        else:
            # 如果没有找到对应的配置，可以设置默认值或抛出异常
            self.chat_stream = False
            self.proxy_url = 'https://llmfw.waf.volces.com/v1/check'
            self.bypass_uri = ''
            self.api_log = '/tmp/'

globalModel = {}


def get_user_config(user_id):
    if user_id not in globalModel:
        globalModel[user_id] = GlobalModel(user_id)
    
    return globalModel[user_id]




def model_management_routes(app):

    #添加模型的参数的初始值
    @app.route('/ai/model/add_model_info', methods=['POST'])
   # @jwt_required()
    def update_model_info():
        try:
            user_id = request.json.get('user_id')
            if not user_id:
                return jsonify({'message': 'user_id不能为空'}), 400
            
            user = get_record_by_field(User, 'id', user_id)
            if not user:
                return jsonify({'message': '用户不存在'}), 404
            
            model_config = get_record_by_field(ModelConfig, 'user_id', user_id)
            if not model_config:  # 如果不存在，则增加模型信息
                model_config = None
                model_config = ModelConfig(
                    user_id=user_id,
                    cdK='xsfghytynhjj',
                    use_stream=False,
                    content_type=1,
                    class_id='',
                    proxy_url='https://llmfw.waf.volces.com/v1/check',
                    bypass_uri='',
                    local_log_path='/tmp/',
                    remote_log_path='',
                    expiry_date='2025-01-01',
                )

            db.session.add(model_config)
            db.session.commit()

            return jsonify({
                'message': '添加成功',
                'code': 200
            })
        
        except Exception as e:
            print(e)
            db.session.rollback()
            app.logger.error(e)
            return jsonify({'message': str(e)}), 500




    #更新模型聊天的参数
    @app.route('/ai/model/update_chat_info', methods=['POST'])
    #@jwt_required()
    def get_chat_info():
        try:
            user_id = request.json.get('user_id')
            if not user_id:
                return jsonify({'message': 'user_id不能为空'}), 400
            
            user = get_record_by_field(User, 'id', user_id)
            if not user:
                return jsonify({'message': '用户不存在'}), 404
            
            model_config = get_record_by_field(ModelConfig, 'user_id', user_id)
            if not model_config:
                return jsonify({'message': '模型不存在'}), 404
            
            model_config.use_stream = request.json.get('use_stream', model_config.use_stream)
            
            db.session.commit()

            return jsonify({
                'message': '更新成功',
                'code': 200
                }), 200
        
        except Exception as e:
            print(e)
            db.session.rollback()
            app.logger.error(e)
            return jsonify({
                'message': str(e),
                'code': 500
                }), 500
        


    #更新模型的网络部署的接口
    @app.route('/ai/model/update_network_info', methods=['POST'])
    #@jwt_required()
    def update_network_info():
        try:
            user_id = request.json.get('user_id')
            model_config_network_status = request.json.get('network_status')

            #model_config_network_status的值必须是0或1
            if model_config_network_status not in [0, 1]:
                return jsonify({'message': 'network_status的值必须是0或1'}), 400
            

            if not user_id:
                return jsonify({'message': 'user_id不能为空'}), 400
            user = get_record_by_field(User, 'id', user_id)
            if not user:
                return jsonify({'message': '用户不存在'}), 404
            

            model_config = get_record_by_field(ModelConfig, 'user_id', user_id)
            if not model_config:
                return jsonify({'message': '模型不存在'}), 404
            
            
            if model_config_network_status == 1:      #1表示串联
                model_config.bypass_uri = request.json.get('bypass_uri', model_config.bypass_uri)   #zhiPUAI
                model_config.config_network_status = model_config_network_status
            else:
                model_config.config_network_status = model_config_network_status
            
            db.session.commit()

            return jsonify({
                'message': '配置成功',
                'code': 200
                }), 200
        
        except Exception as e:
            print(e)
            db.session.rollback()
            app.logger.error(e)
            return jsonify({
                'message': str(e),
                'code': 500
                }), 500




    #更新API日志的存储参数的接口
    @app.route('/ai/model/update_api_log_info', methods=['POST'])
    #@jwt_required()
    def update_api_log_info():
        try:
            user_id = request.json.get('user_id')
            model_api_log_status = request.json.get('api_log_status')

            if not user_id:
                return jsonify({'message': 'user_id不能为空'}), 400
            
            user = get_record_by_field(User, 'id', user_id)
            if not user:
               return jsonify({'message': '用户不存在'}), 404
            
            #model_api_log_status的值必须是0或1
            if model_api_log_status not in [0, 1]:
                return jsonify({'message': 'api_log_status的值必须是0或1'}), 400
            
            model_config = get_record_by_field(ModelConfig, 'user_id', user_id)
            if not model_config:
                return jsonify({'message': '模型不存在'}), 404
            

            if model_api_log_status == 1:     #1表示远程日志
                model_config.remote_log_path = request.json.get('remote_log_path', model_config.remote_log_path)
                model_config.api_log_status = model_api_log_status
            else:
                model_config.local_log_path = request.json.get('local_log_path', model_config.local_log_path)
                model_config.api_log_status = model_api_log_status

            db.session.commit()


            return jsonify({
                'message': '配置成功',
                'code': 200
                }), 200
        
        except Exception as e:
            print(e)
            db.session.rollback()
            app.logger.error(e)
            return jsonify({
                'message': str(e),
                'code': 500
                }), 500


