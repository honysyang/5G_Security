import requests
from flask import Flask, request, Response, jsonify
import os
import json
from datetime import datetime
from database_management import db,DialogLogs,User,DialogDirectoriesUser,get_record_by_field,get_record_by_field_all
from sqlalchemy.exc import SQLAlchemyError
from model_management import get_user_config
from zhipuai import ZhipuAI


def chat_routes(app):
    """chat相关的路由"""


    def AIS_FW(directory_id,user_id,url, network_type,proxy_payload,stat=0, type=0):
        try:
            if proxy_payload['content'] == '':
                return None, 0, 0
            
            if network_type == 1:  #表示串联部署
                proxy_payload['use_stream'] = False
                response = requests.post(url, json=proxy_payload)

                response_data = response.json()   #将AI的响应内容传输给AIS大模型

                # 解析response
                base_resp = response_data["base_resp"]
                decision = response_data["decision"]
                suggest = response_data["suggest"]

                current_id = 0
                current_timestamp = ""

                if type == 0:
                    print(f"user输入: {proxy_payload['content']}\n")
                elif type == 2 and stat == 0:
                    print(f"智普AI响应: {proxy_payload['content']}\n")

                    new_AI_dialogLogs = DialogLogs(
                            user_id=user_id,
                            directory_id=directory_id,
                            output_content=proxy_payload['content'],
                            status_code=base_resp['status_code'],
                            status_message=base_resp['status_message'],
                            err_code=decision['err_code'],
                            err_msg=decision['err_msg'],
                            labels = decision['labels'],
                            type = type
                        )

                    db.session.add(new_AI_dialogLogs)
                    db.session.flush()  # 刷新session以获取新用户的ID

                    # 提交事务
                    db.session.commit()
                    current_id = new_AI_dialogLogs.id

                    current_timestamp = new_AI_dialogLogs.timestamp

                if decision['err_code'] == 0:  #表示输入内容安全
                    stat = 0; suggest = proxy_payload['content']
                else: 
                    stat = 1;
                
                if stat == 1:
                    print(f"AIS检测: {suggest}\n")
                    new_AIS_dialogLogs = DialogLogs(
                            user_id=user_id,
                            directory_id=directory_id,
                            output_content= suggest,
                            status_code=base_resp['status_code'],
                            status_message=base_resp['status_message'],
                            err_code=decision['err_code'],
                            err_msg=decision['err_msg'],
                            labels = decision['labels'],
                            type = 3
                        )
                    

                    db.session.add(new_AIS_dialogLogs)
                    db.session.flush()  # 刷新session以获取新用户的ID

                    #获取当前插入数据的id
                    current_id = new_AIS_dialogLogs.id

                    current_timestamp = new_AIS_dialogLogs.timestamp
                    

                return suggest,stat,type+1,current_id ,current_timestamp     #type=1,表示是用户输入   ,type=2,表示是AI输入，stat =0表示正常，stat=1表示异常

                
            else:  #表示旁路部署
                #此时可以根据proxy_payload['use_stream']判断是否流式输出
                response = requests.post(url, json=proxy_payload)
                if response.status_code != 200:
                    return None, 0, 0

                response_data = response.json()   #将AI的响应内容传输给AIS大模型

                decision = response_data["decision"]
                suggest = response_data["suggest"]

                if decision['err_code'] == 0:  #表示输入内容安全
                    stat = 0;
                else: 
                    stat = 1;
                
                return suggest,stat,type+1      #type=1,表示是用户输入   ,type=2,表示是AI输入，stat =0表示正常，stat=1表示异常
                
        except Exception as e:
            print(e)
            return None, 0, 0


    def AI_chat(url, type,dest_payload):
        try:
            if dest_payload['content'] == '':
                return None, 0

            
            dest_payload['use_stream'] = False


            #client = ZhipuAI(api_key="729760d9e110526baa5b5bc52cedbbee.FS1Q0i6PBAKR3nth") # 请填写您自己的APIKey

            #response = requests.post(url, json=dest_payload)

            input = dest_payload['content']


            client = ZhipuAI(api_key="729760d9e110526baa5b5bc52cedbbee.FS1Q0i6PBAKR3nth") # 请填写您自己的APIKey
            response = client.chat.completions.create(
                model="glm-4",  # 填写需要调用的模型名称
                messages=[
                    {"role": "user", "content": input}
                ],
            )
            message_dict = {
                'content': response.choices[0].message.content,
                'role': response.choices[0].message.role,
                # 如果tool_calls有数据，也要加入到字典中
                # 'tool_calls': response.choices[0].message.tool_calls,
            }

            # 使用jsonify返回JSON响应
            return message_dict['content'],type+1

        
        except Exception as e:
            print(e)
            return None, 0


    @app.route('/ai/chat', methods=['POST'])
    def handle_chat():
        # 检查请求中是否包含必要的字段
        if 'content' not in request.json:
            return jsonify({'error': 'Missing required fields'}), 400
        
        
        # 获取请求参数
        input = request.json['content']
        user_id = request.json.get('user_id')
        directory_id = request.json.get('directory_id')

        user_config = get_user_config(user_id)

        
        # 调用chat接口的URL
        if user_config:          
            network_type = user_config.config_network_status
            use_stream = user_config.chat_stream
            proxy_url = user_config.proxy_url
            dest_url = user_config.bypass_uri

        else:
            network_type = 0
            use_stream = request.json.get('use_stream', '')
            proxy_url = 'https://llmfw.waf.volces.com/v1/check'
            dest_url = 'https://open.bigmodel.cn/api/paas/v4/chat/completions'
        

        # 准备请求数据
        proxy_payload = {
                        "token": "1VzZuxtbaK1L5o72",
                        "content": input,
                        "use_stream": use_stream
                    }


        dest_payload = {
                        "token": "729760d9e110526baa5b5bc52cedbbee.FS1Q0i6PBAKR3nth",
                        "content": input,
                        "use_stream": False
                    }

        #定义变量
        type = 0   #定义初始值
        stat = 0   #定义初始值
        output = '' #定义初始值
        current_id = 0
        current_timestamp = ""
    

        try:

            if directory_id:
                directory = get_record_by_field(DialogDirectoriesUser, 'id', directory_id)
            
            else:
                directory_name = input[:10]
                directory = get_record_by_field(DialogDirectoriesUser, 'directory_name', directory_name)

            directory_id = directory.id

            #将输入部分内容存储到数据库中
            input_dialogLogs = DialogLogs(
                                user_id=user_id,
                                directory_id=directory_id,
                                input_content=input,
                                type = type     #user
                            )
            db.session.add(input_dialogLogs)
            db.session.commit()  # 直接提交，避免使用flush以简化事务管理

             

            if network_type == 1:   #Ture表示串联部署
                output,stat,type ,current_id, current_timestamp= AIS_FW(directory_id,user_id,proxy_url, network_type,proxy_payload,stat, type)
                if stat == 0:
                    dest_payload['content'] = output
                    
                    output = ''
                    output,type = AI_chat(dest_url, type,dest_payload)
                

                    proxy_payload['content'] = output
                    output,stat,type,current_id, current_timestamp = AIS_FW(directory_id,user_id,proxy_url, network_type,proxy_payload,stat, type)
                
                
                if stat == 1:
                    output = "[AIS网关:]"+" "+output
                else:
                    output = "[AI:    ]"+" "+output

                return jsonify({
                        "message": "chat successfully.",
                        "code":200,
                        "data":{
                            "directory_id":directory_id,
                            "id":current_id,
                            "user_id":user_id,
                            "current_timestamp":current_timestamp,
                            "suggest":output
                        }
                    }), 201

            else:   #False表示并联部署
                use_stream = proxy_payload['use_stream']
                # 根据use_stream的值选择不同的响应处理方式
                if use_stream:
                    # 流式输出处理
                    def generate():
                        
                        first_base_resp = None  # 用于累积base_resp数据
                        first_decision_data = None  # 用于累积decision数据

                        response = requests.post(proxy_url, json=proxy_payload, stream=True)
                        buffer = b""
                        last_suggest = ""
                        end_status = False

                        # 构建新的事件数据，包括额外的元数据
                        # 构建初始的事件数据，包括额外的元数据
                        event_data = {
                            "code": 200,
                            "data": {
                                "user_id": user_id,
                                "directory_id": directory_id,
                                "suggest": "",  
                                "labels": "",
                                "err_msg": "",
                                "end_status": end_status
                            }
                        }             

                        for chunk in response.iter_content(chunk_size=None):
                            if chunk:

                                # 立即发送'suggest'字段的值给客户端
                                #yield chunk.decode('utf-8')
                                
                                # 将新接收到的数据添加到缓冲区
                                buffer += chunk

                                # 查找完整的事件，它们以\n\n分隔
                                events = buffer.split(b'\n\n')
                                buffer = events.pop()  # 最后一个可能不完整，保留在buffer中

                                # 遍历每一个完整的事件
                                for event in events:
                                    lines = event.split(b'\n')
                                    data_line = next((line for line in lines if line.startswith(b'data:')), None)
                                    
                                    if data_line:
                                        # 解析data行中的JSON数据
                                        data_json = json.loads(data_line[5:].decode('utf-8'))  # 去掉'data:'前缀
                                        suggest = data_json.get('suggest', '')

                                        # 累积'suggest'字段
                                        last_suggest += suggest

                                        # 更新base_response和decision_data，如果存在的话
                                        if first_base_resp is None and 'base_resp' in data_json:
                                            first_base_resp = data_json['base_resp']
                                        if first_decision_data is None and 'decision' in data_json:
                                            first_decision_data = data_json['decision']


                                    # 构建新的事件数据，包括额外的元数据
                                    event_data["data"].update({
                                        "suggest": last_suggest,
                                        "labels": first_decision_data['labels'],
                                        "err_msg": first_decision_data['err_msg'],
                                        "end_status": end_status
                                    })

                                    # 发送包含元数据的事件
                                    yield f"data: {json.dumps(event_data)}\n\n"


                        end_status = True
    

                        # 在流结束时，处理缓冲区中的任何剩余数据
                        if buffer:
                            lines = buffer.split(b'\n')
                            data_line = next((line for line in lines if line.startswith(b'data:')), None)
                            if data_line:
                                data_json = json.loads(data_line[5:].decode('utf-8'))
                                suggest = data_json.get('suggest', '')
                                last_suggest += suggest

                        # 在流结束后存储数据到数据库
                        try:
                            with app.app_context():  # 确保在应用上下文中执行数据库操作
                                new_dialogLogs = DialogLogs(
                                    user_id=user_id,
                                    directory_id=directory_id,
                                    output_content=last_suggest,  # 使用累积的suggest
                                    status_code=first_base_resp['status_code'] if first_base_resp else 0,
                                    status_message=first_base_resp['status_message'] if first_base_resp else '',
                                    err_code=first_decision_data['err_code'] if first_decision_data else 0,
                                    err_msg=first_decision_data['err_msg'] if first_decision_data else '',
                                    labels=first_decision_data['labels'] if first_decision_data else '',
                                    type = 3
                                )
                                db.session.add(new_dialogLogs)
                                db.session.commit()  # 直接提交，避免使用flush以简化事务管理

                                #获取当前插入数据的id
                                current_id = new_dialogLogs.id
                                current_timestamp = new_dialogLogs.timestamp

                                event_data["data"].update({
                                    "id": current_id,
                                    "timestamp": current_timestamp.isoformat(),
                                    "end_status": end_status
                                })

                                # 最后再发送包含元数据的事件
                                yield f"data: {json.dumps(event_data)}\n\n"

                        except SQLAlchemyError as e:
                            db.session.rollback()  # 回滚事务以处理可能的数据库错误
                            app.logger.error(f"Database Error: {str(e)}")

                    return Response(generate(),content_type='text/event-stream'),200
                
                else:
                    # 非流式输出处理
                    response = requests.post(proxy_url, json=proxy_payload)
                    if response.status_code != 200:
                        return jsonify({'error': 'Failed to fetch response from chat service'}), 500

                    response_data = response.json()

                    # 解析response
                    base_resp = response_data["base_resp"]
                    decision = response_data["decision"]
                    suggest = response_data["suggest"]


                    new_dialogLogs = DialogLogs(
                        user_id=user_id,
                        directory_id=directory_id,
                        output_content=suggest,
                        status_code=base_resp['status_code'],
                        status_message=base_resp['status_message'],
                        err_code=decision['err_code'],
                        err_msg=decision['err_msg'],
                        labels = decision['labels'],
                        type = 3
                    )


                    db.session.add(new_dialogLogs)
                    db.session.flush()  # 刷新session以获取新用户的ID

                    # 提交事务
                    db.session.commit()

                    #获取当前插入数据的id
                    current_id = new_dialogLogs.id

                    current_timestamp = new_dialogLogs.timestamp
                    message = new_dialogLogs.output_content


                    if decision['err_msg'] == 'pass':
                        message = "当前系统检测策略是PASS"

                    message = "[AIS检测:] "+ message

                    return jsonify({
                        "message": "chat successfully.",
                        "code":200,
                        "data":{
                            "directory_id":directory_id,
                            "id":current_id,
                            "user_id":user_id,
                            "current_timestamp":current_timestamp,
                            "suggest":message
                        }
                    }), 201


        except Exception as e:
            db.session.rollback()  # 发生异常时回滚事务
            app.logger.error(f"Error adding chat: {str(e)}")  # 记录错误日志
            return jsonify({"error": "An error occurred while adding the chat."}), 500



   
    # 创建对话目录
    @app.route('/ai/directories/create', methods=['POST'])
    def create_directory():
        try:
            user_id = request.json.get('user_id')
            # 获取请求参数
            content = request.json['content'] 
            
            user = get_record_by_field(User, 'id', user_id)
            if not user:
                return jsonify({"error": "User not found."}), 404
                
            name = content[:10]
            
            directory = DialogDirectoriesUser(directory_name=name, user_id=user_id)
            db.session.add(directory)
            db.session.commit()

            return jsonify({
                "message": "Directory created successfully.",
                "data": {
                    "directory_id": directory.id,
                    "directory_name": directory.directory_name,
                    "user_id": directory.user_id,
                    "created_at": directory.created_at.strftime("%Y-%m-%d %H:%M:%S")

                },
                "code":200
                }), 201
            
        except Exception as e:
            db.session.rollback()
            return jsonify({"message": "An error occurred while creating the directory."}), 500



    #修改对话目录的名称
    @app.route('/ai/directories/update', methods=['PUT'])
    #@jwt_required()
    def update_directories():
        try:
            directory_id = request.json.get('directory_id')
            directory_name = request.json.get('directory_name')
            user_id = request.json.get('user_id')
            if not directory_id or not directory_name or not user_id:
                return jsonify({'message': '参数错误','code':400}), 400
            #user_id = get_jwt_identity()
            directory = DialogDirectoriesUser.query.filter_by(id=directory_id, user_id=user_id).first()
            if directory:
                directory.directory_name = directory_name
                db.session.commit()
                return jsonify({'status': 'success', 'message': '目录名称修改成功'})
            else:
                return jsonify({'status': 'error', 'message': '目录不存在'})
            
        except Exception as e:
            return jsonify({'status': 'error', 'message': str(e)})
    


    # 删除对话目录
    @app.route('/ai/directories/delete', methods=['DELETE'])
    #@jwt_required()
    def delete_dialog_directory():
        try:
            #user_id = get_jwt_identity()
            
            user_id = request.json.get('user_id')
            directory_id = request.json.get('directory_id')
            directory = DialogDirectoriesUser.query.filter_by(id=directory_id, user_id=user_id).first()

            if directory:
                db.session.delete(directory)

                dialogLogs = get_record_by_field_all(DialogLogs,'directory_id',directory_id)
                for log in dialogLogs:
                    db.session.delete(log)

                db.session.commit()
                return jsonify({
                    'status': 'success', 
                    'message':'目录删除成功',
                    'code':200
                    })
            return jsonify({
                'status': 'error', 
                'message':'目录为空',
                'code':400
                })
        except Exception as e:
            print(e)
            return jsonify({
                'status': 'error', 
                'message':'异常操作',
                'code':400
                }) 




# 删除对话内容
    @app.route('/ai/directories/clear', methods=['DELETE'])
    #@jwt_required()
    def clear_content_directory():
        try:
            #user_id = get_jwt_identity()
            
            user_id = request.json.get('user_id')
            directory_id = request.json.get('directory_id')
            

            dialogLogs = get_record_by_field_all(DialogLogs,'directory_id',directory_id)
            for log in dialogLogs:
                db.session.delete(log)

                db.session.commit()
            return jsonify({
                'status': 'success', 
                'message':'对话内容删除成功',
                'code':200
            })
            
        except Exception as e:
            print(e)
            return jsonify({
                'status': 'error', 
                'message':'异常操作',
                'code':400
                }) 



    # 获取对话目录列表
    @app.route('/ai/directories/all', methods=['POST'])
    #@jwt_required()
    def get_directories():
        user_id = request.json.get('user_id')
        try:
            directories = DialogDirectoriesUser.query.filter_by(user_id=user_id).order_by(DialogDirectoriesUser.id.desc()).all()
            if directories:
                directorie_list = [
                {
                    'id': directorie.id,
                    'directory_name': directorie.directory_name
                }
                for directorie in directories
                ]

                return jsonify({
                    'status': 'success', 
                    'message':'目录获取成功',
                    'code':200,
                    'data': directorie_list
                    })
            else:
                return jsonify({
                    'status': 'error', 
                    'message':'当前目录为空',
                    'code':200,
                    'data': []
                    })
        except Exception as e:
            print(e)
            return jsonify({
                'status': 'error', 
                'message':'异常操作',
                'code':400,
                'data': []
                })




    #点击聊天目录，显示该目录下的对话记录
    @app.route('/ai/directories/chat/log', methods=['POST'])
    #@jwt_required()
    def get_dialog_by_directory():
        try:
            directory_id = request.json.get('directory_id')
            user_id = request.json.get('user_id')
            dialog_logs = DialogLogs.query.filter_by(directory_id=directory_id, user_id=user_id).all()
            dialog_logs_list = []
            
            for dialog_log in dialog_logs:
                log_type = "user" if dialog_log.input_content else "ai"  # Determine the type based on input_content
                content = dialog_log.input_content if dialog_log.input_content else dialog_log.output_content  # Decide the content
                dialog_logs_list.append({
                    'id': dialog_log.id,
                    'type': log_type,
                    'content': content,
                    'timestamp': dialog_log.timestamp.strftime('%Y-%m-%d %H:%M:%S')
                })
            return jsonify({
                'code': 200,
                'msg': 'success',
                'data': {
                    'dialog_logs': dialog_logs_list
                }
            })
        except Exception as e:
            return jsonify({
                'code': 500,
                'msg': 'error',
                'data': {
                    'error': str(e)
                }
            })


