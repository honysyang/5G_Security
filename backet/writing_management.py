from database_management import db,ContactUs,AboutUs,Image
from flask import Flask, request, jsonify


#定义文案管理的路由
def writing_route(app):

    #获取文案列表的信息
    @app.route('/ai/writing_management/get_writing_list', methods=['POST'])
    def get_writing_list():
        try:
            #展示ContactUs表的数据
            contact_us_list = ContactUs.query.all()

            contact_us_lists = []

            for contact_us in contact_us_list:
                contact_us_lists.append({
                    'id': contact_us.id,
                    'contact_info': contact_us.contact_info,
                    'image_path': contact_us.image_path,
                    'created_at': contact_us.created_at,
                    'updated_at': contact_us.updated_at
                })
            

            #展示AboutUs表的数据
            about_us_list = AboutUs.query.all()
            about_us_lists = []
            for about_us in about_us_list:
                about_us_lists.append({
                    'id': about_us.id,
                    'title': about_us.title,
                    'content': about_us.content,
                    'created_at': about_us.created_at,
                    'updated_at': about_us.updated_at
                })

            return jsonify({
                'data':{
                    'contact_us_list': contact_us_lists,
                    'about_us_list': about_us_lists
                },
                'code': 200,
                'success': True
            })
        
        except Exception as e:
            return jsonify({
                'code': 500,
                'success': False,
                'message': str(e)
            })


            
    #更新关于我们的内容
    @app.route('/ai/writing_management/update_about_us', methods=['PUT'])
    def update_about_us():
        try:
            #获取前端传递的参数
            title = request.form.get('title')
            content = request.form.get('content')
            id = request.form.get('about_id')
            if not title or not content or not id:
                return jsonify({
                    'code': 400,
                    'message': '参数错误'
                })
            
            about_us = AboutUs.query.filter_by(id=id).first()
            if not about_us:
                return jsonify({
                    'code': 404,
                    'message': 'about_id错误'
                })
            
            about_us.title = title
            about_us.content = content

            db.session.commit()
            return jsonify({
                'code': 200,
                'message': '更新成功'
            })
        
        except Exception as e:
            db.session.rollback()
            app.logger.error(f"Error updating about_us: {str(e)}")
            return jsonify({
                'code': 500,
                'message': str(e)
            })
            
    #更新联系我们表的内容
    @app.route('/ai/writing_management/update_contact_us', methods=['PUT'])
    def update_contact_us():
        try:
            #获取前端传递的参数
            id = request.form.get('contact_id')
            image_path = request.form.get('image_path')
            contact_info = request.form.get('contact_info')

            if not id or not image_path or not contact_info:
                return jsonify({
                    'code': 400,
                    'msg': '参数错误'
                })
            
            contact_us = ContactUs.query.filter_by(id=id).first()
            if not contact_us:
                return jsonify({
                    'code': 400,
                    'msg': 'contact_id错误'
                })
            
            contact_us.image_path = image_path
            contact_us.contact_info = contact_info

            #同步更新到Image表中，type=3
            image = Image.query.filter_by(type=3).first()
            if image:
                image.path = image_path
                image.status = 1

            db.session.commit()
            return jsonify({
                'code': 200,
                'message': '更新成功'
            })
        
        except Exception as e:
            db.session.rollback()
            app.logger.error(f"Error updating about_us: {str(e)}")
            return jsonify({
                'code': 500,
                'message': str(e)
            })
            
