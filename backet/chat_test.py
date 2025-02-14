import requests
from flask import Flask, request, Response, jsonify
import os
import json
from datetime import datetime
from database_management import db,DialogLogs,User,DialogDirectoriesUser,get_record_by_field,get_record_by_field_all
from sqlalchemy.exc import SQLAlchemyError
from model_management import get_user_config
from zhipuai import ZhipuAI


def chat_test_routes(app):
    """chat相关的路由"""

    def AIS_FW(url, network_type,proxy_payload,stat=0, type=0):
        try:
            if proxy_payload['content'] == '':
                return None, 0, 0
            
            if network_type == 1:  #表示串联部署
                proxy_payload['use_stream'] = False
                response = requests.post(url, json=proxy_payload)

                response_data = response.json()   #将AI的响应内容传输给AIS大模型

                decision = response_data["decision"]
                suggest = response_data["suggest"]

                if type == 0:
                    print(f"user输入: {proxy_payload['content']}\n")
                elif type == 2 and stat == 0:
                    print(f"智普AI响应: {proxy_payload['content']}\n")


                if decision['err_code'] == 0:  #表示输入内容安全
                    stat = 0; suggest = proxy_payload['content']
                else: 
                    stat = 1;

                
                if stat == 1:
                    print(f"AIS检测: {suggest}\n")
                

                return suggest,stat,type+1      #type=1,表示是用户输入   ,type=2,表示是AI输入，stat =0表示正常，stat=1表示异常

                
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



    @app.route('/ai/chat_test', methods=['POST'])
    def test_chat():
            # 检查请求中是否包含必要的字段
        if 'content' not in request.json:
            return jsonify({'error': 'Missing required fields'}), 400         
            
        # 获取请求参数
        input = request.json['content']
        use_stream = request.json.get('use_stream', '')

        # 调用chat接口的URL
        proxy_url = 'https://llmfw.waf.volces.com/v1/check'

        #智普AI
        dest_url = 'https://open.bigmodel.cn/api/paas/v4/chat/completions'
            
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

        network_type = request.json.get('network_type', 1)   #假设从前端读取

        
        if network_type == 1:   #Ture表示串联部署
            output,stat,type = AIS_FW(proxy_url, network_type,proxy_payload,stat, type)
            if stat == 0:
                dest_payload['content'] = output
                
                output = ''
                output,type = AI_chat(dest_url, type,dest_payload)
            

                proxy_payload['content'] = output
                output,stat,type = AIS_FW(proxy_url, network_type,proxy_payload,stat, type)
            
            return jsonify({'output': output}), 200
        

        else:   #False表示并联部署
            output, stat,type = AIS_FW(proxy_url, type, network_type,proxy_payload, stat) 
            return jsonify({'output': output}), 200


            
    @app.route('/ai/chat_ai_test', methods=['POST'])
    def test_ai_chat():
            # 检查请求中是否包含必要的字段
        if 'content' not in request.json:
            return jsonify({'error': 'Missing required fields'}), 400         
            
        # 获取请求参数
        input = request.json['content']

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

        print(f"suggest:{message_dict['content']}\n")
        # 使用jsonify返回JSON响应
        return jsonify(message_dict)

















