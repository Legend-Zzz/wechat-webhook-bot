import json
import os
import requests
from django.http import JsonResponse
from django.shortcuts import render
from django.views.decorators.csrf import csrf_exempt

# alertmanager发送webhook(post请求)，经过本程序处理，再将处理后的数据post到企业微信bot机器人接口，进而发送告警通知
def send_alert(receive_data):
    try:
        webhook_url = os.environ.get('WXWORK_WEBHOOK_BOT_URL')
        if not webhook_url:
            return JsonResponse({"error": "未设置 WXWORK_WEBHOOK_BOT_URL 环境变量"}, status=400)

        alerts = json.loads(receive_data.decode('utf-8'))
        
        firing_alerts = [alert for alert in alerts['alerts'] if alert['status'] == 'firing']
        resolved_alerts = [alert for alert in alerts['alerts'] if alert['status'] == 'resolved']
        
        firing_msgs = "\r\n".join([alert['annotations']['description'] for alert in firing_alerts])
        resolved_msgs = "\r\n".join([alert['annotations']['description'] for alert in resolved_alerts])
        
        firing_count = len(firing_alerts)
        resolved_count = len(resolved_alerts)
        
        if firing_count > 0 and resolved_count > 0:
            data = f"[{firing_count}]  未恢复的告警\n{firing_msgs}\n[{resolved_count}]  已恢复的告警\n{resolved_msgs}"
        elif firing_count > 0 and resolved_count == 0:
            data = f"[{firing_count}]  未恢复的告警\n{firing_msgs}"
        elif firing_count == 0 and resolved_count > 0:
            data = f"[{resolved_count}]  已恢复的告警\n{resolved_msgs}"
        else:
            data = 'error, no data'

        # 发送告警消息
        send_data = {"msgtype": "text", "text": {"content": data}}
        headers = {"Content-Type": "application/json"}
        
        response = requests.post(url=webhook_url, headers=headers, json=send_data)
        response.raise_for_status()
        
        return response.text
        
    except Exception as e:
        return JsonResponse(f"发生错误: {e}", status=500)

# GET请求返回主页，POST请求调用send_alert方法
@csrf_exempt
def index(request):
    if request.method == 'GET':
        return render(request, 'index.html')
    elif request.method == 'POST':
        try:
            receive_data = request.body
            send_alert(receive_data)
            return JsonResponse({"message": "告警通知已发送"})
        except Exception as e:
            return JsonResponse({"error": f"发生错误: {e}"}, status=500)
