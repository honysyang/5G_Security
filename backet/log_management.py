from database_management import db,DialogLogs,Role,User,UserRole
from flask import Flask, request, jsonify
import datetime
from sqlalchemy import and_, or_
from sqlalchemy.sql import func



def log_routes(app):
    #日志管理的接口
    @app.route('/ai/logs/management/display', methods=['POST'])
    def log_display():
        try:
            # 获取用户输入的参数
            request_data = request.get_json()

            user_id = request_data.get('user_id', '')
            label = request_data.get('label', '')
            action = request_data.get('action', '')  # 新增的action参数
            keyword = request_data.get('keyword', '')
            page = request_data.get('page', 1)
            page_size = request_data.get('page_size', 10)
            sort = request_data.get('sort', 'desc')
            
            # 处理时间参数
            time_range = request_data.get('time', [])
            start_time = time_range[0] if len(time_range) > 0 else None
            end_time = time_range[1] if len(time_range) > 1 else None

            # 构建查询基础
            query = DialogLogs.query

            # 添加过滤条件
            if user_id:
                query = query.filter(DialogLogs.user_id == user_id)
            if label:
                query = query.filter(DialogLogs.labels.like(f'%{label}%'))
            if keyword:
                query = query.filter(or_(DialogLogs.input_content.like(f'%{keyword}%'), DialogLogs.output_content.like(f'%{keyword}%')))
            if start_time and end_time:
                query = query.filter(and_(DialogLogs.timestamp >= start_time, DialogLogs.timestamp <= end_time))
            if action:
                # 根据action过滤err_msg
                if action.lower() == 'allow':
                    query = query.filter(DialogLogs.err_msg.like('%allow%'))
                elif action.lower() == 'block':
                    query = query.filter(DialogLogs.err_msg.like('%block%'))

            # 添加排序
            if sort == 'desc':
                query = query.order_by(DialogLogs.timestamp.desc())
            else:
                query = query.order_by(DialogLogs.timestamp.asc())

            # 分页
            offset = (page - 1) * page_size
            results = query.offset(offset).limit(page_size).all()
            total_records = query.count()


            logs = []
            for result in results:
                    content = result.output_content if result.output_content else result.input_content
                    object = 'USER' if result.input_content else 'AIS'
                    log = {
                        'log_id': result.id,
                        'user_id': result.user_id,
                        'object': object,
                        'content': content,
                        'timestamp': result.timestamp,
                        'action': result.err_msg,
                        'labels': result.labels
                    }
                    logs.append(log)

            # 返回结果
            return jsonify({
                'data':{
                    'list':logs,
                    'total_records': total_records
                },
                'code':200,
                'success': True
            })

        except Exception as e:
            return jsonify({'error': str(e)}), 500
    

    @app.route('/ai/logs/management/info', methods=['POST'])
    def log_info():
        try:
            # 获取用户输入的参数
            request_data = request.get_json()
            user_id = request_data.get('user_id', '')
            log_id = request_data.get('log_id', '')

            if not user_id or not log_id:
                return jsonify({'error': 'user_id and log_id are required'}), 400
            
            # 基于条件获得日志记录
            log = DialogLogs.query.filter_by(user_id=user_id, id=log_id).first()

            if not log:
                return jsonify({'error': 'Log not found'}), 404
            
            #获得用户信息
            user = User.query.filter_by(id=user_id).first()
            if not user:
                return jsonify({'error': 'User not found'}), 404


            # 获得角色信息
            userrole = UserRole.query.filter_by(user_id=user_id).first()
            if not userrole:
                return jsonify({'error': 'UserRole not found'}), 404

            role = Role.query.filter_by(id=userrole.role_id).first()
            if not role:
                return jsonify({'error': 'Role not found'}), 404

            user_role = role.alias  # 用户角色信息


            object = 'USER' if log.input_content else 'AIS'
            length = len(log.input_content) if log.input_content else len(log.output_content)

            # 构造返回的JSON数据
            response_data = {
                'log': {
                    'user_id': log.user_id,
                    'object': object,
                    'directory_id': log.directory_id,
                    'input_content': log.input_content,
                    'output_content': log.output_content,
                    'length': length,
                    'timestamp': log.timestamp.strftime('%Y-%m-%d %H:%M:%S'),
                    'status_code': log.status_code,
                    'status_message': log.status_message,
                    'err_code': log.err_code,
                    'err_msg': log.err_msg,
                    'labels': log.labels,
                    'type': log.type
                },
                'user_role': user_role,
                "user_info" : {
                    'id': user.id,
                    'nickname': user.nickname,
                    'account': user.account,
                    'avatar_url': user.avatar_url,
                    'phone_number': user.phone_number,
                    'status': user.status,
                    'created_at': user.created_at.strftime('%Y-%m-%d %H:%M:%S')
                }
            }

            # 返回结果
            return jsonify({
                'data': response_data,
                'code': 200,
                'success': True
            })

        except Exception as e:
            return jsonify({'error': str(e)}), 500


