
"""
主要功能如下：
1. 解析靶场环境运行模板
2. 根据运行模板生成对应的任务
3. 根据任务生成需要运行的shell脚本
4. 运行shell脚本
"""
    

import json
import logging
from typing import Dict, Any

# 设置日志
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
log_file = 'execution.log'

class TaskGenerator:
    def __init__(self, template_file: str):
        self.template_file = template_file
        self.template_data = None

    def load_template(self) -> bool:
        """加载并解析JSON模板文件"""
        try:
            with open(self.template_file, 'r') as file:
                self.template_data = json.load(file)
            logging.info("Template loaded successfully.")
            return True
        except FileNotFoundError:
            logging.error(f"Template file not found: {self.template_file}")
            return False
        except json.JSONDecodeError as e:
            logging.error(f"Failed to parse JSON template: {e}")
            return False
        
    def generate_task(self) -> Dict[str, Any]:
        """基于解析的模板生成任务"""

        task = {
            'image': self.template_data['image'],
            'environment': self.template_data['environment'],
            'volumes': self.template_data['volumes'],
            'network': self.template_data['network'],
            'resources': self.template_data['resources'],
            'command': self.template_data['command']
        }

        logging.info("Task generated successfully.")
        return task
    
    def generate_shell_script(self, task: Dict[str, Any], script_file: str) -> bool:
        """生成启动容器的Shell脚本"""
        try:
            with open(script_file, 'w') as file:
                file.write("#!/bin/bash\n")
                file.write(f"docker run --rm ")

                # 添加环境变量
                for key, value in task['environment'].items():
                    file.write(f"-e {key}={value} ")

                # 添加卷映射
                for volume in task['volumes']:
                    file.write(f"-v {volume['host_path']}:{volume['container_path']}:{volume['mode']} ")

                # 添加网络配置
                if task['network']['mode'] == 'host':
                    file.write("--network host ")
                elif task['network']['mode'] == 'none':
                    file.write("--network none ")
                else:
                    for host_port, container_port in task['network']['ports'].items():
                        file.write(f"-p {host_port}:{container_port} ")

                # 添加资源限制
                if 'cpu' in task['resources']:
                    file.write(f"--cpus={task['resources']['cpu']} ")
                if 'memory' in task['resources']:
                    file.write(f"--memory={task['resources']['memory']} ")

                # 添加镜像和命令
                file.write(f"{task['image']['name']}:{task['image']['tag']} ")
                file.write(' '.join(task['command']))
                file.write("\n")

            logging.info(f"Shell script generated successfully: {script_file}")
            return True
        except Exception as e:
            logging.error(f"Failed to generate shell script: {e}")
            return False
        
    
    def execute_shell_script(self, script_file: str) -> bool:
        """执行生成的Shell脚本"""
        try:
            result = subprocess.run(['bash', script_file], stdout=log, stderr=log, text=True)
            logging.info(f"Script executed successfully:\n{result.stdout.decode()}")
            return True
        except subprocess.CalledProcessError as e:
            logging.error(f"Script execution failed:\n{e.stderr.decode()}")
            return False
        
        
    
    