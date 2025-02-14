from database_management import DialogLogs,db
from flask import Flask, request, jsonify
import datetime
from sqlalchemy import and_, or_
from sqlalchemy.sql import func
import logging
from collections import Counter
import re


def flow_routes(app):
    #流量统计的接口
    @app.route('/ai/flows/management/display', methods=['POST'])
    def flow_display():
        try:
            print("标签统计")

            query = db.session.query(
            DialogLogs.labels,
            db.func.count(DialogLogs.id).label('count')
            ).filter(
            DialogLogs.labels != None,  # 过滤掉None值
            DialogLogs.labels != ''  # 过滤掉空字符串
            ).group_by(
            DialogLogs.labels
            ).order_by(
            db.desc('count')  # 按count降序排序
            )

            # 执行查询并获取结果
            results = query.all()

            if results:
                logging.info("labels 查询并获取结果")

            # 初始化一个字典来统计每个单独标签的计数
            label_counts = {}

            # 初始化两个列表来分别存储标签和对应的计数
            labels_list = []
            counts_list = []
 
 
            # 遍历查询结果，并拆分每个逗号分隔的标签
            for result in results:
                labels = result.labels.split(',')
                for label in labels:
                    label = label.strip()  # 去除可能的空白字符
                    if label:  # 忽略空字符串
                        label_counts[label] = label_counts.get(label, 0) + result.count
        
            # 将字典转换为两个列表：标签和对应的计数
            labels_list = list(label_counts.keys())
            counts_list = [label_counts[label] for label in labels_list]
            
            # 格式化结果，使得标签和对应的计数是平行的两个列表
            formatted_label_results = {'label': labels_list, 'count': counts_list}

            print("动作统计")
            
            query_err_msg = db.session.query(
                DialogLogs.err_msg,
                db.func.count(DialogLogs.id).label('count')
            ).filter(
                DialogLogs.err_msg != None  # 过滤掉NULL值
            ).group_by(
                DialogLogs.err_msg
            ).order_by(
                db.desc('count')  # 按count降序排序
            )
 

            # 执行查询并获取结果
            results_err_msg = query_err_msg.all()

            # 初始化两个列表来分别存储错误消息和对应的计数
            err_msgs = []
            counts = []
     
            # 遍历查询结果，将数据添加到列表中
            for result in results_err_msg:
                err_msgs.append(result.err_msg)
                counts.append(result.count)

            err_msg_counts = {'err_msg': err_msgs, 'count': counts}


            print("input_content 统计")

            query_input_content = db.session.query(DialogLogs.input_content)

            results_input_content = query_input_content.all()

            if results_input_content:
                print('获取频率最高的几个词')
 
            # 将所有input_content文本合并到一个大字符串中
            text_data_input_content = ' '.join([result.input_content for result in results_input_content if result.input_content])
        
            # 使用正则表达式来分割单词
            words_input_content = re.findall(r'\w+', text_data_input_content.lower())

            # 确保单词列表中不包含None
            words_input_content = [word for word in words_input_content if word]
        
            # 计算单词频率
            word_counts_input_content = Counter(words_input_content)
        
            # 获取频率最高的几个词
            most_common_words_input_content = word_counts_input_content.most_common(15)  # 假设我们要找出前10个最频繁的词

            print('获取频率最高的几个词')

            word_cloud_input_content_data = {word: count for word, count in most_common_words_input_content}
        
            # 格式化结果
            formatted_results_input_content = [{'word': word, 'count': count} for word, count in most_common_words_input_content]



            print("output_content 统计")

            query_output_content = db.session.query(DialogLogs.output_content)

            results_output_content = query_output_content.all()

            if results_output_content:
                print('获取频率最高的几个词')
 
            # 将所有input_content文本合并到一个大字符串中
            text_data_output_content = ' '.join([result.output_content for result in results_output_content if result.output_content])
        
            # 使用正则表达式来分割单词
            words_output_content = re.findall(r'\w+', text_data_output_content.lower())

            # 确保单词列表中不包含None
            words = [word for word in words_output_content if word]
        
            # 计算单词频率
            word_counts = Counter(words)
        
            # 获取频率最高的几个词
            most_common_words = word_counts.most_common(15)  # 假设我们要找出前10个最频繁的词

            print('获取频率最高的几个词')

            word_cloud_output_content_d123456ata = {word: count for word, count in most_common_words}

            response_data = {
                'label_counts_data': formatted_label_results,
                'err_msg_counts_data': err_msg_counts,
                'input_content_data' : word_cloud_input_content_data,
                'output_content_data' : word_cloud_output_content_data
            }
        
            # 返回JSON格式的响应
            return jsonify({"data":response_data, 'success': True, 'code':200 }), 200


        except Exception as e:
            logging.error(f"message:{e}")
            return jsonify({'message': str(e)}), 500
