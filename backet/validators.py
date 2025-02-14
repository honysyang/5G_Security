import re

def validate_nickname(nickname):
    """验证昵称是否合法。"""
    if not isinstance(nickname, str) or not re.match(r'^[\w\u4e00-\u9fa5]{1,10}$', nickname):
        raise ValueError("Nickname must be 1 to 10 characters, including Chinese, English letters, digits, and underscores.")

def validate_account(account):
    """验证账号是否符合格式。"""
    if not isinstance(account, str) or not re.match(r'^[a-zA-Z0-9]{8,12}$', account):
        raise ValueError("Account must be 8 to 12 alphanumeric characters.")

def validate_avatar_url(avatar_url):
    """验证头像URL是否安全。"""
    if avatar_url and not avatar_url.startswith(('http://', 'https://')):
        raise ValueError("Avatar URL must start with http:// or https://.")

def validate_phone_number(phone_number):
    """验证电话号码是否只包含数字且格式合理。"""
    if phone_number and not re.match(r'^\d{8,15}$', phone_number):
        raise ValueError("Phone number must contain only digits and can range from 8 to 15 digits.")
def validate_password(password):
    """验证密码复杂度，包含大小写字母、数字和特殊字符,还有#也合法。"""
    if not re.match(r'^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,12}$', password):
        raise ValueError("Password must be 8 to 12 characters long and contain at least one uppercase letter, one lowercase letter, one digit, and one special character (@$!%*?&).")

def validate_status(status):
    """验证状态是否合法。"""
    if status not in [0, 1]:
        raise ValueError("Status must be either 0 or 1.")
    
def validate_role_name(role_name):
    """验证role_name不能出现恶意字符和空格、引号等"""
    if not isinstance(role_name, str) or not re.match(r'^[^<>\"\'\\/]*$', role_name):
        raise ValueError("Role name cannot contain <>, \", ' or \\ or /.")
    

def validate_remark(remark):
    """验证备注不能出现恶意字符和空格、引号等"""
    if not isinstance(remark, str) or not re.match(r'^[^<>\"\'\\/]*$', remark):
        raise ValueError("Remark cannot contain <>, \", ' or \\ or /.")