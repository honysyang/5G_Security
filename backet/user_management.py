import os
from database_management import init_db, db,User,Role,execute_query,UserRole,RoleMenu,Menu,ModelConfig,Image,DialogDirectoriesUser,DialogLogs
from auth_manament import init_jwt, register_routes,jwt_required
from flask import Flask, request, jsonify
from validators import validate_password,validate_nickname,validate_account,validate_avatar_url,validate_phone_number,validate_status,validate_role_name,validate_remark
from werkzeug.security import generate_password_hash, check_password_hash
from sqlalchemy.exc import IntegrityError
from sqlalchemy import or_


#这里属于用户管理和角色管理模块
def user_management_routes(app):
    """用户和角色管理相关的路由"""


    #查看用户管理信息
    @app.route('/ai/user_management/admin/display_user', methods=['POST'])
    def get_users():
        """获取所有用户列表的API接口"""
        try:
            size = request.json.get('size', 10)
            page = request.json.get('page', 1)

            keyword = request.json.get('keyword',"")

            if keyword:
                users = User.query.filter(or_(User.nickname.like(f'%{keyword}%'), User.account.like(f'%{keyword}%'))).all()
                serialized_users = [
                    {
                        'user_id': user.id,
                        'nickname': user.nickname,
                        'account': user.account,
                        'avatar_url': user.avatar_url,
                        'phone_number': user.phone_number,
                        'role_id':UserRole.query.filter_by(user_id = user.id).first().role_id,
                        'status': user.status,
                        'created_at': user.created_at.isoformat()
                    }
                    for user in users
                ]

                return jsonify({
                    "message": "User search successful.",
                    "code": 200,
                    "data": serialized_users,
                    "success": True
                })

            # 计算偏移量
            offset = (page - 1) * size

            # 查询用户总数
            total_query = "SELECT COUNT(*) FROM users"
            total = execute_query(total_query)[0][0]

            # 查询用户列表，包括关联角色名称
            # 注意：这里简化处理，假设有一个简单的JOIN查询可以获取到角色名称，实际中可能需要根据数据库设计调整
            users_query = f"""
                SELECT u.id, u.nickname, u.account, u.avatar_url, u.phone_number, r.name AS role_name,r.id AS role_id, u.status, u.created_at
                FROM users u
                LEFT JOIN user_role ur ON u.id = ur.user_id
                LEFT JOIN roles r ON ur.role_id = r.id
                LIMIT {size} OFFSET {offset}
            """
            users = execute_query(users_query)

            # 序列化用户数据
            serialized_users = [
                {
                    'user_id': record[0],
                    'nickname': record[1],
                    'account': record[2],
                    'avatar_url': record[3],  # 处理可能的None值
                    'phone_number': record[4],
                    'role_name':record[5],
                    'role_id': record[6] ,  # 获取角色名称，处理NULL
                    'status': record[7],
                    'created_at': record[8].isoformat() if record[7] else None,
                }
                for record in users
            ]
            # 准备响应数据
            response_data = {
                "success": True,
                "message": "Users fetched successfully",
                "data": {
                    "total": total,
                    "list": serialized_users
                },
                "code":200
            }

            return jsonify(response_data), 200

        # 错误处理部分也应遵循统一格式
        except Exception as e:
            app.logger.error(f"Error adding user: {str(e)}")  # 记录错误日志
            error_response = {
                "success": False,
                "message": "An error occurred while fetching users",
                "error": str(e)
            }
            return jsonify(error_response), 500



    #展示角色管理中的角色信息，用于增加用户时，显示下拉
    @app.route('/ai/user_management/admin/display_alias', methods=['POST'])
    def get_alias():
        """获取所有用户列表的API接口"""
        try:
            # 查询角色表中的所有别名
            query = "SELECT id,name FROM roles"
            names = execute_query(query)

            # 序列化角色别名
            serialized_aliases = [{"id": record[0],'name': record[1]} for record in names]

            # 统一响应格式
            response_data = {
                "success": True,
                "message": "Role aliases fetched successfully",
                "data": serialized_aliases,
                "code":200
            }

            return jsonify(response_data), 200

        except Exception as e:
            # 统一错误响应格式
            error_response = {
                "success": False,
                "message": "An error occurred while fetching role aliases",
                "error": str(e)
            }
            return jsonify(error_response), 500
            


    #增加用户
    @app.route('/ai/user_management/admin/add_user', methods=['POST'])
    def add_user():
        """增加用户"""
        try:
            data = request.get_json()
            nickname = data.get('nickname',"")     #默认为空
            account = data.get('account')
            avatar_url = data.get('avatar_url', "")  # 设置默认值为空字符串
            phone_number = data.get('phone_number', "")   #默认为空
            role_id = data.get('role_id')
            status = data.get('status', 1)  # 默认启用状态

            # 确保角色名称存在并获取其ID
            role = Role.query.filter_by(id=role_id).first()
            if not role:
                return jsonify({"error": "Role not found"}), 404

            # 检查用户是否已存在
            existing_user = User.query.filter((User.nickname == nickname) | (User.account == account)).first()
            if existing_user:
                return jsonify({"error": "User already exists."}), 400

            # 密码加密处理
            app.config['PASSWORD'] = os.getenv('PASSWORD')
            password = generate_password_hash(app.config['PASSWORD'])

            # 创建用户实体
            new_user = User(
                nickname=nickname,
                account=account,
                avatar_url=avatar_url,
                phone_number=phone_number,
                password=password,
                status=status
            )
            db.session.add(new_user)
            db.session.flush()  # 刷新session以获取新用户的ID

            # 创建用户角色关联
            user_role = UserRole(user_id=new_user.id, role_id=role.id)
            db.session.add(user_role)

            # 提交事务
            db.session.commit()

            return jsonify({
                "message": "User added successfully.",
                "code":200,
                "success": True
                }), 201

        except Exception as e:
            db.session.rollback()  # 发生异常时回滚事务
            app.logger.error(f"Error adding user: {str(e)}")  # 记录错误日志
            return jsonify({"error": "An error occurred while adding the user."}), 500



    #用户密码重置
    @app.route('/ai/user_management/admin/reset_password', methods=['POST'])
    def reset_password():
        try:
            data = request.get_json()
            
            reset_password = data.get('reset_password_status')

            user_ids_str = data.get('user_ids')     #"1,2,3,4"
            
            
            user_ids = [int(id.strip()) for id in user_ids_str.split(',')]

            print(user_ids)

            if reset_password == 1:
                for user_id in user_ids:
                    # 检查用户是否存在
                    user = User.query.get(user_id)
                    if not user:
                        return jsonify({"error": "User not found."}), 404

            
                    #从.env中获取PASSWORD
                    app.config['PASSWORD'] = os.getenv('PASSWORD')
                    user.password = generate_password_hash(app.config['PASSWORD'])
                    user.init_pwd_status = 0

            else:
                return jsonify({
                    "message": "reset_password error.",
                    "code":200,
                    "success": True
                }), 200

            db.session.commit()
            return jsonify({
                "message": "Password reset successfully.",
                "code":200,
                "success": True
            }), 200
        
        except Exception as e:
            db.session.rollback()  # 发生异常时回滚事务
            app.logger.error(f"Error resetting password: {str(e)}")  # 记录错误日志
            return jsonify({"error": "An error occurred while resetting the password."}), 500



    #删除用户
    @app.route('/ai/user_management/admin/delete_user', methods=['DELETE'])
    def delete_user():
        """删除用户"""
        try:
            # 获取请求中的数据
            data = request.get_json()
            user_ids_str = data.get('user_ids')     #"1,2,3,4"
            
            
            user_ids = [int(id.strip()) for id in user_ids_str.split(',')]

            print(user_ids)

            for user_id in user_ids:
                user = User.query.get(user_id)
                # 检查用户是否存在
                if not user:
                    return jsonify({"error": "User not found."}), 404

                # 删除用户角色关联
                UserRole.query.filter_by(user_id=user_id).delete()

                #删除模型参数信息
                ModelConfig.query.filter_by(user_id=user_id).delete()
                '''
                Image.query.filter_by(user_id=user_id).delete()

                DialogDirectoriesUser.filter_by(user_id=user_id).delete()

                DialogLogs.filter_by(user_id=user_id).delete()
                '''

                # 删除用户
                db.session.delete(user)

            db.session.commit()

            return jsonify({
                "message": "User deleted successfully.",
                "code": 200,
                "success": True
            }), 200

            
        except Exception as e:
            db.session.rollback()  # 发生异常时回滚事务
            app.logger.error(f"Error deleting user: {str(e)}")  # 记录错误日志
            return jsonify({"error": "An error occurred while deleting the user."}), 500
        


    #编辑用户信息
    @app.route('/ai/user_management/admin/update_user_info', methods=['PUT'])
    def update_user_info():
        """修改用户"""
        try:
            # 获取请求中的数据
            data = request.get_json()
            user_id = data.get('user_id')
            account = data.get('account')
            role_id = data.get('role_id')
            nickname = data.get('nickname')
            status = data.get('status',1)

            # 检查用户是否存在
            user = User.query.get(user_id)
            if not user:
                return jsonify({"error": "User not found."}), 404
            
            if status == 1:
                user.status = 1
            else:
                user.status = 0

            if account:
                user.account = account

            if nickname:
                user.nickname = nickname
                
            # 角色名称转ID
            if role_id:
                role = Role.query.filter_by(id=role_id).first()
                if not role:
                    return jsonify({"error": "Role not found"}), 404
                # 确保关联表的更新逻辑正确，这里简化处理，实际可能需要先删除旧关联再添加新关联
                user_role = UserRole.query.filter_by(user_id=user_id).first()
                if user_role:
                    user_role.role_id = role.id
                else:
                    user_role = UserRole(user_id=user_id, role_id=role.id)
                    db.session.add(user_role)
            
            

            # 提交更改
            db.session.commit()
            
            return jsonify(message="User status updated successfully.",
                           code = 200,
                           success = True
                           ), 200
        except Exception as e:
            db.session.rollback()
            app.logger.error(f"Error updating user status: {str(e)}")
            return jsonify(500, description=f"An error occurred: {str(e)}")



    #更新用户的状态
    @app.route('/ai/user_management/admin/update_user_status', methods=['PUT'])
    def update_user_status():
        """修改用户"""
        try:
            # 获取请求中的数据
            data = request.get_json()
            user_id = data.get('user_id')
            status = data.get('status',1)

            # 检查用户是否存在
            user = User.query.get(user_id)
            if not user:
                return jsonify({"error": "User not found."}), 404
            
            if status == 1:
                user.status = 1
            else:
                user.status = 0

            # 提交更改
            db.session.commit()
            
            return jsonify(message="User status updated successfully.",
                           code = 200,
                           success = True
                           ), 200
        except Exception as e:
            db.session.rollback()
            app.logger.error(f"Error updating user status: {str(e)}")
            return jsonify(500, description=f"An error occurred: {str(e)}")
        




    #更新用户角色信息
    @app.route('/ai/user_management/admin/update_user_role', methods=['PUT'])
    def update_user():
        """修改用户记录"""
        try:
            # 获取请求中的数据
            data = request.get_json()

            user_ids_str = data.get('user_ids')     #"1,2,3,4"
            
            
            user_ids = [int(id.strip()) for id in user_ids_str.split(',')]

            print(user_ids)
             # 更新用户信息，注意检查并处理密码更新逻辑
            role_id = data.get('role_id')

            for user_id in user_ids:
                # 查询用户是否存在
                user = User.query.get(user_id)
                if not user:
                    return jsonify({"error": "User not found"}), 404

                # 角色名称转ID
                if role_id:
                    role = Role.query.filter_by(id=role_id).first()
                    if not role:
                        return jsonify({"error": "Role not found"}), 404
                    # 确保关联表的更新逻辑正确，这里简化处理，实际可能需要先删除旧关联再添加新关联
                    user_role = UserRole.query.filter_by(user_id=user_id).first()
                    if user_role:
                        user_role.role_id = role.id
                    else:
                        user_role = UserRole(user_id=user_id, role_id=role.id)
                        db.session.add(user_role)
                else:
                    return jsonify({"error": "role_name not found"}), 404


            db.session.commit()

            return jsonify({
                "message": "User updated successfully.",
                "code": 200,
                "success": True
            }), 200

        except Exception as e:
            db.session.rollback()
            app.logger.error(f"Error updating user: {str(e)}")
            return jsonify({"error": "An error occurred while updating the user."}), 500
        



    #查看角色管理信息
    @app.route('/ai/role_management/all_roles', methods=['POST'])
    def get_all_roles():
        """获取所有角色信息的API接口，支持分页，使用execute_query查询"""
        try:
            # 获取分页参数
            size = request.json.get('size', 10)
            page = request.json.get('page', 1)

            keyword = request.json.get('keyword',"")

            if keyword:
                roles = Role.query.filter(Role.name.like(f'%{keyword}%')).all()
                serialized_roles = [
                    {
                        'role_id': role.id,
                        'name': role.name,
                        'alias': role.alias,  # 假设Role模型中有alias属性
                        'created_at': role.created_at.isoformat(),
                        'status': role.status,
                        'remark': role.remark,  # 假设Role模型中有remark属性
                        'menus_ids':role.menus_ids,
                        'menus_tree':role.menus_tree       #role.menus_tree = menus_tree
                    }
                    for role in roles
                ]
                return jsonify({
                    "message": "Role search successful.",
                    "code": 200,
                    "data": serialized_roles,
                    "success": True
                })

            # 计算偏移量
            offset = (page - 1) * size

            # 查询角色总数
            count_query = "SELECT COUNT(*) FROM roles"
            total = execute_query(count_query)[0][0]

            # 分页查询角色数据
            roles_query = f"""
                SELECT id, name, alias, created_at, status, remark,menus_ids,menus_tree
                FROM roles
                ORDER BY id
                LIMIT {size} OFFSET {offset}
            """
            roles = execute_query(roles_query)

            # 序列化角色数据
            serialized_roles = [
                {
                    'role_id': role.id,
                    'name': role.name,
                    'alias': role.alias,  # 假设Role模型中有alias属性
                    'created_at': role.created_at.isoformat(),
                    'status': role.status,
                    'remark': role.remark,  # 假设Role模型中有remark属性
                    'menus_ids':role.menus_ids,
                    'menus_tree':role.menus_tree
                }
                for role in roles
            ]

            # 构建统一响应格式，包含分页信息和总数
            response_data = {
                "success": True,
                "message": "Roles fetched successfully",
                "data": {
                    "total": total,
                    "list": serialized_roles    
                },
                "code":200
            }

            return jsonify(response_data), 200

        except Exception as e:
            # 统一错误响应格式
            error_response = {
                "success": False,
                "message": "An error occurred while fetching roles",
                "error": str(e)
            }
            return jsonify(error_response), 500
        
    


    # 创建角色
    @app.route('/ai/user_management/admin/create_role', methods=['POST'])
    def create_role():
        """创建角色的API接口"""
        try:
            # 获取请求中的数据
            data = request.get_json()
            name = data.get('name')
            alias = data.get('alias')
            remark = data.get('remark', "")  # 默认为空字符串
            status = data.get('status', 1)  # 默认启用状态

            # 检查alias是否已存在
            existing_role = Role.query.filter_by(alias=alias).first()
            if existing_role:
                return jsonify({"error": "Role alias already exists."}), 400

            # 创建角色实体
            new_role = Role(name=name, alias=alias, remark=remark, status=status)
            db.session.add(new_role)

            # 尝试提交到数据库，捕获唯一性约束错误（例如alias重复）
            try:
                db.session.commit()

                return jsonify({
                    "success": True,
                    "message": "Role created successfully.",
                    "code": 200
                }), 201  # 使用201 Created状态码
            except IntegrityError as e:
                db.session.rollback()
                # 根据错误类型判断是否是因为alias重复，这里简化处理，实际情况可能需要更细致的错误分析
                return jsonify({
                    "success": False,
                    "error": "Role alias already exists.",
                    "code": 400}), 400

        except Exception as e:
            db.session.rollback()  # 发生异常时回滚事务
            app.logger.error(f"Error creating role: {str(e)}")  # 记录错误日志
            return jsonify({"error": "An error occurred while creating the role."}), 500
    


    #修改角色的状态
    @app.route('/ai/user_management/admin/update_roles_status', methods=['PUT'])
    def update_roles_status():
        """修改用户"""
        try:
            # 获取请求中的数据
            data = request.get_json()
            role_id = data.get('role_id')
            status = data.get('status',1)


            # 查询角色是否存在
            role = Role.query.get(role_id)
            if not role:
                return jsonify({"error": "Role not found"}), 404
            
            
            role.status = status

            # 提交更改到数据库
            db.session.commit()
            
            return jsonify(message="User status updated successfully.",
                           success= True,
                           code = 200
                           ), 200
        except Exception as e:
            return jsonify(500, description=f"An error occurred: {str(e)}")
        



    #角色分配菜单
    @app.route('/ai/user_management/admin/role_menu_assign', methods=['PUT'])
    def role_menu_assign():
        """角色分配菜单"""
        try:
            # 获取请求中的数据
            data = request.get_json()
            role_id = data.get('role_id')
            menu_ids_str = data.get('menu_ids')       # menu_ids 的样子是“1，2，3，4”
            menus_tree = data.get('menu_tree')


            # 将字符串转换为整数列表
            menu_ids = [(id.strip()) for id in menu_ids_str.split(',')]

            for menu_id in menu_ids:
            # 检查是否已经存在相同的role_id和menu_id组合
                existing_role_menu = RoleMenu.query.filter_by(role_id=role_id, menu_id=menu_id).first()
                if not existing_role_menu:
                    # 创建关联实体
                    role_menu = RoleMenu(role_id=role_id, menu_id=menu_id)
                    db.session.add(role_menu)
            
            role = Role.query.get(role_id)
            role.menus_ids = menu_ids_str
            role.menus_tree = menus_tree

            db.session.commit()
            return jsonify(
                message="Role menu assigned successfully.",
                code = 200,
                success = True
            ), 200
        
        except Exception as e:
            db.session.rollback()
            app.logger.error(f"Error assigning role menu: {str(e)}")
            return jsonify(
                code = 500, 
                description=f"An error occurred: {str(e)}"
                )
            
            





    # 更新角色记录
    @app.route('/ai/user_management/admin/update_role', methods=['PUT'])
    def update_role():
        """更新角色信息的API接口"""
        try:
            # 获取请求中的数据
            data = request.get_json()
            role_id = data.get('role_id')
            
            # 查询角色是否存在
            role = Role.query.get(role_id)
            if not role:
                return jsonify({"error": "Role not found"}), 404

            # 更新角色信息
            if 'name' in data:
                role.name = data['name']
            if 'alias' in data:
                # 需要确保alias的唯一性，这里简化处理，实际应先检查alias是否已存在
                role.alias = data['alias']
            if 'remark' in data:
                role.remark = data.get('remark', "")
            if 'status' in data:
                role.status = data.get('status', 1)

            # 提交更改到数据库
            db.session.commit()

            return jsonify({
                "message": "Role updated successfully.",
                "code": 200,
                "success": True
                }), 200

        except Exception as e:
            db.session.rollback()  # 发生异常时回滚事务
            app.logger.error(f"Error updating role: {str(e)}")  # 记录错误日志
            return jsonify({"error": "An error occurred while updating the role."}), 500
        
    



    # 删除角色记录
    @app.route('/ai/user_management/admin/delete_role', methods=['DELETE'])
    def delete_role():
        """删除角色的API接口"""
        try:
            # 获取请求中的数据
            data = request.get_json()
            role_id = data.get('role_id')

            # 查询角色是否存在
            role = Role.query.get(role_id)
            if not role:
                return jsonify({"error": "Role not found"}), 404
            
            # 判断角色是否被用户绑定
            if UserRole.query.filter_by(role_id=role_id).count() > 0:
                user_roles = UserRole.query.filter_by(role_id=role_id).all()
                for user_role in user_roles:
                    db.session.delete(user_role)
            

            #判断角色是否被菜单绑定，如果存在，则删除对应的RoleMenu表中的数据
            if RoleMenu.query.filter_by(role_id=role_id).count() > 0:
                role_menus = RoleMenu.query.filter_by(role_id=role_id).all()
                for role_menu in role_menus:
                    db.session.delete(role_menu)
                    
            # 删除角色表中的角色
            db.session.delete(role)
            
            db.session.commit()

            return jsonify({
                "message": "Role deleted successfully.",
                "code": 200,
                "success": True}), 200
            
        except Exception as e:
            db.session.rollback()  # 发生异常时回滚事务
            app.logger.error(f"Error deleting role: {str(e)}")  # 记录错误日志
            return jsonify(500, description=f"An error occurred: {str(e)}")
    





    #查询菜单
    @app.route('/ai/menu_management/admin/menus', methods=['POST'])
    def get_menus():
        """获取所有菜单项，支持通过角色ID筛选角色有权访问的菜单"""
        data = request.get_json()
        role_id = data.get('role_id', None)
        try:
            # 如果提供了角色ID，则查询该角色有权访问的菜单
            if role_id:
                # 使用关联查询，获取角色对应的菜单
                query = db.session.query(Menu).join(RoleMenu, Menu.id == RoleMenu.menu_id).filter(RoleMenu.role_id == role_id)
            else:
                # 否则，查询所有菜单
                query = Menu.query

            # 执行查询并构造响应数据
            menus = query.all()
            menu_list = [
                {
                    "id": menu.id,
                    "parentId": menu.parent_id,
                    "name": menu.name,
                    "title": menu.title,
                    "path": menu.path,
                    "component": menu.component,
                    "icon": menu.icon,
                    "orderNum": menu.order_num,
                    "status": menu.status,
                    "createdAt": menu.created_at.strftime('%Y-%m-%d %H:%M:%S'),
                    "updatedAt": menu.updated_at.strftime('%Y-%m-%d %H:%M:%S')
                }
                for menu in menus
            ]
                        
            response_data = {
                "success": True,
                "code": 200,
                "message": "fetch menus successfully",
                "data": menu_list,
            }

            return jsonify(response_data), 200
        except Exception as e:
            return jsonify({"error": "Failed to fetch menus.", "description": str(e)}), 500



    # 创建菜单
    @app.route('/ai/menu_management/admin/create_menu', methods=['POST'])
    def create_menu():
        """创建菜单项,支持通过角色ID关联角色与菜单"""
        try:
            # 获取请求中的数据
            data = request.get_json()
            parent_id = data.get('parentId', None)
            name = data.get('name', None)
            title = data.get('title', None)
            path = data.get('path', None)
            component = data.get('component', None)
            icon = data.get('icon', None)
            order_num = data.get('orderNum', 0)
            status = data.get('status', 'ACTIVE').upper()  # 确保状态为大写
            
        
            
            # 创建新的菜单项
            new_menu = Menu(
                parent_id=parent_id,
                name=name,
                title=title,
                path=path,
                component=component,
                icon=icon,
                order_num=order_num,
                status=status
            )
            
            db.session.add(new_menu)
            db.session.flush()  # flush to get the new_menu's id before committing
            
            # 如果请求中包含了角色ID，关联角色与菜单
            
            role_id = data.get('role_id')
            if role_id:
                role_menu_association = RoleMenu(role_id=role_id, menu_id=new_menu.id)
                db.session.add(role_menu_association)
            
            db.session.commit()

            data = {
                "menuId": new_menu.id,
                "menu": {
                    "id": new_menu.id,
                    "parentId": new_menu.parent_id,
                    "name": new_menu.name,
                    "title": new_menu.title,
                    "path": new_menu.path,
                    "component": new_menu.component,
                    "icon": new_menu.icon,
                    "orderNum": new_menu.order_num,
                    "status": new_menu.status
                    }
            }
            
            return jsonify({
                "message": "Menu item created successfully.",
                "code":200,
                "data":new_menu.id
            }), 201

        except IntegrityError as e:
            # 处理唯一性约束冲突等数据库错误
            db.session.rollback()
            app.logger.error(f"Error create menu: {str(e)}")  # 记录错误日志
            return jsonify({"error": "Database integrity error.", "description": str(e),"code":409}), 409  # 409 Conflict
        
        except Exception as e:
            db.session.rollback()
            app.logger.error(f"Error create menu: {str(e)}")  # 记录错误日志
            return jsonify({"error": "An unexpected error occurred.", "description": str(e),"code":500}), 500



    #更新菜单
    @app.route('/ai/menu_management/admin/update_menu', methods=['PUT'])
    def update_menu():
        """更新菜单项"""
        try:
            # 获取请求中的数据
            data = request.get_json()
            menu_id = data.get('id')
            parent_id = data.get('parentId')

            
            # 更新菜单信息
            name = data.get('name', '').strip()
            title = data.get('title', '').strip()
            path = data.get('path', None)
            component = data.get('component', None)
            icon = data.get('icon', None)
            order_num = data.get('orderNum', 0)
            status = data.get('status', 'ACTIVE').upper()  # 确保状态为大写

            # 查询菜单是否存在
            menu = db.session.query(Menu).filter_by(id=menu_id,parent_id = parent_id).first()
            

            # 更新菜单信息
            if name:
                menu.name = name
            if title:
                menu.title = title
            if path is not None:
                menu.path = path
            if component is not None:
                menu.component = component
            if icon is not None:
                menu.icon = icon
            if order_num != 0:
                menu.order_num = order_num
            if status in ('ACTIVE', 'INACTIVE'):
                menu.status = status

            # 处理角色关联更新（如果请求中包含角色ID）
            role_id = data.get('roleId',None)
            if role_id is not None:
                # 先删除旧的关联
                RoleMenu.query.filter_by(menu_id=menu_id).delete()
                # 添加新的关联
                if role_id:
                    role_menu_association = RoleMenu(role_id=role_id, menu_id=menu_id)
                    db.session.add(role_menu_association)

            db.session.commit()

            return jsonify({
                "message": "Menu item updated successfully.",
                "code": 200,
                "success": True
                })

        except IntegrityError as e:
            db.session.rollback()
            app.logger.error(f"Error update menu: {str(e)}")  # 记录错误日志
            return jsonify({"error": "Database integrity error.", "description": str(e)}), 409  # 409 Conflict

        except Exception as e:
            db.session.rollback()
            app.logger.error(f"Error update menu: {str(e)}")  # 记录错误日志
            return jsonify({"error": "An unexpected error occurred.", "description": str(e)}), 500
        




    @app.route('/ai/menu_management/admin/menu_display',methods=['POST'])
    def menu_detail():
        try:
            data = request.get_json()
            menu_id = data.get('menu_id')

            menu = db.session.query(Menu).filter_by(id=menu_id).first()
            
            # Serialize the menu object into JSON format
            menu_data = {
                "id": menu.id,
                "parent_id": menu.parent_id,
                "name": menu.name,
                "title": menu.title,
                "path": menu.path,
                "component": menu.component,
                "icon": menu.icon,
                "order_num": menu.order_num,
                "status": menu.status,
                "created_at": menu.created_at.isoformat(),
                "updated_at": menu.updated_at.isoformat()
            }
            
            return jsonify({
                "message": "Menu item fetched successfully.",
                "code":200,
                "data":menu_data
            }), 201

            
        except Exception as e:
            db.session.rollback()
            app.logger.error(f"Error deleting menu: {str(e)}")  # 记录错误日志
            return jsonify({"error": "An unexpected error occurred.", "description": str(e)}), 500





    # 删除菜单项
    @app.route('/ai/menu_management/admin/delete_menu', methods=['DELETE'])
    def delete_menu():
        """删除菜单项，同时处理与之关联的子菜单和角色关联"""
        try:
            data = request.get_json()
            ids_str = data.get('menu_ids', '')  # 从前端获取id序列字符串
            ids_list = [int(id) for id in ids_str.split(',') if id.isdigit()]  # 将字符串转换为整数列表


            for id in ids_list:
                menu = Menu.query.get(id)
                if not menu:
                    continue

                try:
                    db.session.delete(menu)
                    
                except Exception as e:
                    db.session.rollback()
                    

            db.session.commit()

            return jsonify({
                "message": "Menu item and its associations have been deleted successfully.",
                "code": 200,
                "success": True
                }), 200

        except IntegrityError as e:
            db.session.rollback()
            app.logger.error(f"Error deleting menu: {str(e)}")  # 记录错误日志
            return jsonify({"error": "Database integrity error.", "description": str(e)}), 409  # 409 Conflict

        except Exception as e:
            db.session.rollback()
            app.logger.error(f"Error deleting menu: {str(e)}")  # 记录错误日志
            return jsonify({"error": "An unexpected error occurred.", "description": str(e)}), 500




            
                    
            
            