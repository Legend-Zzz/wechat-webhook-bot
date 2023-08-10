import json
import os

import requests
from django.http import HttpResponse
from django.shortcuts import render


# Create your views here.

# alertmanager发送webhook(post请求)，经过本程序处理，再将处理后的数据post到企业微信bot机器人接口，进而发送告警通知
def send_alert(receive_data):
    headers = {"Content-Type": "application/x-www-form-urlencoded"}
    url = os.environ.get('WXWORK_WEBHOOK_BOT_URL')

    receive_data = receive_data.decode('utf8').replace("'", '"')
    receive_data = json.loads(receive_data)

    firing_count = resolved_count = 0
    firing_msg = resolved_msg = ''
    for i in receive_data['alerts']:
        if i['status'] == 'firing':
            firing_count += 1
            firing_msg = '\r\n'.join([firing_msg, i['annotations']['description']])
        elif i['status'] == 'resolved':
            resolved_count += 1
            resolved_msg = '\r\n'.join([resolved_msg, i['annotations']['description']])

    if firing_count > 0 and resolved_count > 0:
        data = '''[{0}]  未恢复的告警
{1}
[{2}]  已恢复的告警
{3}'''.format(firing_count, firing_msg.strip(), resolved_count, resolved_msg.strip())
    elif firing_count > 0 and resolved_count == 0:
        data = '''[{0}]  未恢复的告警
{1}'''.format(firing_count, firing_msg.strip())
    elif firing_count == 0 and resolved_count > 0:
        data = '''[{0}]  已恢复的告警
{1}'''.format(resolved_count, resolved_msg.strip())
    else:
        data = 'error, no data'

# 发送告警消息
    send_data = {"msgtype":"text","text":{"content":data}}
    send_data = json.dumps(send_data)

    req = requests.post(url=url, headers=headers, data=send_data)
    return req


# GET请求返回主页，POST请求调用send_alert方法
def index(request):
    if request.method == 'GET':
        return render(request, 'index.html')
    else:
        receive_data = request.body
        result = send_alert(receive_data)
        return HttpResponse(result)
