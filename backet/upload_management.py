import os
from database_management import  db, Image,get_record_by_field,get_record_by_field_all,User,ContactUs
from werkzeug.utils import secure_filename
from flask import Flask, request, jsonify



def upload_image_routes(app):

    # 设置上传文件夹和允许的文件扩展名
    UPLOAD_FOLDER = os.getenv('UPLOAD_FOLDER')
    ALLOWED_EXTENSIONS = os.getenv('ALLOWED_EXTENSIONS')
    MAX_CONTENT_LENGTH = os.getenv('MAX_CONTENT_LENGTH')


    '''
        UPLOAD_FOLDER = '/static/upload'
        ALLOWED_EXTENSIONS = {'png'}
        MAX_CONTENT_LENGTH = 16 * 1024 * 1024
    '''


    # 检查扩展名是否允许
    def allowed_file(filename):
        return '.' in filename and \
            filename.rsplit('.', 1)[1].lower() in ALLOWED_EXTENSIONS


    # 参数验证装饰器
    def validate_request(required_fields=('user_id', 'file')):
        def decorator(func):
            def wrapper(*args, **kwargs):
                data = request.form.to_dict()  # 获取表单数据
                files = request.files
                for field in required_fields:
                    if field not in data and field != 'file':
                        return jsonify({'message': f'{field} is required'}), 400
                    elif field == 'file' and 'file' not in files:
                        return jsonify({'message': 'No file part'}), 400
                return func(*args, **kwargs)
            return wrapper
        return decorator


    # 图片上传路由
    @app.route('/ai/image/upload', methods=['PUT'])
    def upload_image():
        data = request.form.to_dict()
        user_id = data.get('user_id')
        image_type = int(data.get('type', 1))  # 默认类型为1
        file = request.files['file']

        if not user_id:
            return jsonify({'message': 'user_id is required'}), 400
        
        #判断用户是否是真实的
        if not get_record_by_field(User, 'user_id', user_id):  # 检查用户是否存在
            return jsonify({'message': 'User not found'}), 404

         # 检查文件名是否为空
        if file.filename == '':
            return jsonify({'message': 'No selected file'}), 400
        
        # 文件类型检查
        if not allowed_file(file.filename):
            return jsonify({'message': 'File type not allowed', 'allowed_types': ALLOWED_EXTENSIONS}), 400
        
        # 文件安全命名
        filename = secure_filename(file.filename)
        
        # 文件大小限制
        if file.content_length > MAX_CONTENT_LENGTH:
            return jsonify({'message': 'File too large', 'max_size': MAX_CONTENT_LENGTH}), 400
        
        try:
            # 文件路径
            file_path = os.path.join(UPLOAD_FOLDER, filename)
            
            # 保存文件
            file.save(file_path)

            #判断image中是否以及存在类似的记录，通过user_id、type来判断,如果存在类似的记录，则选用更新，如果不存在，则新增该记录
            # 查询现有记录
            existing_image = Image.query.filter_by(user_id=user_id, type=image_type).first()

            if image_type == '5':
                #图片也存储到ContactUs表中,表中只有一个列，所以直接更新
                existing_contactus = ContactUs.query.first()
                existing_contactus.image_path = file_path

            if existing_image:
                # 更新现有记录
                existing_image.filename = filename
                existing_image.path = file_path
                db.session.commit()
                return jsonify({
                    'message': 'Image updated successfully',
                    'image_id': existing_image.id,
                    'code': 200
                })
            else:
                # 创建新记录
                new_image = Image(filename=filename, path=file_path, type=image_type, user_id=user_id)
                db.session.add(new_image)
                db.session.commit()
                return jsonify({
                    'message': 'Image uploaded successfully',
                    'image_id': new_image.id,
                    'code': 200
                })
        except Exception as e:
            return jsonify({
                'message': 'Error uploading image', 
                'code':500,
                'error': str(e)
            }), 500
        




    #图片查看详情
    @app.route('/ai/image/display', methods=['POST'])
    def display_image():
        data = request.get_json()
        user_id = data.get('user_id')
         
        if not user_id:
            return jsonify({'message': 'User ID is required'}), 400
        
        images = Image.query.filter_by(user_id = user_id).all()
        
        images_list = []

        for image in images:
            images_list.append({
                'image_id': image.id,
                'filename': image.filename,
                'path': image.path,
                'type': image.type,
                'user_id': image.user_id,
                'created_at': image.created_at.strftime('%Y-%m-%d %H:%M:%S')
            })

        
        return jsonify({
            'message': 'Image found',
            'code':200,
            'data': images_list
            })
        