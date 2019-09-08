import os
import requests
import json
import time
import random

'''
Sending message to Slack
'''


# shortcut for sending noti to slack
def slack_notification(status, options, output=None):
    if not options.get('SLACK'):
        return
    if output:
        slack_file('report', options, mess={
            'title':  "{0} | {1} | Output".format(options['TARGET'], options['CURRENT_MODULE']),
            'filename': '{0}'.format(output),
        })
    else:
        slack_noti(str(status), options, mess={
            'title':  "{0} | {1}".format(options['TARGET'], options['CURRENT_MODULE']),
            'content': 'Start {1} for {0}'.format(options['TARGET'], options['CURRENT_MODULE'])
        })


# Slack printing
def slack_seperate(options):
    if options['BOT_TOKEN'] == "None":
        return
    mess = {
        'title': "Just Seperated stuff",
        'color': "#44475A",
        'image_url': get_emoji(),
        'icon': get_emoji(),
    }
    sm = Messages(options)

    mess['channel'] = options['STATUS_CHANNEL']
    sm.send_mess(mess)
    mess['channel'] = options['LOG_CHANNEL']
    sm.send_mess(mess)
    mess['channel'] = options['REPORT_CHANNEL']
    sm.send_mess(mess)
    mess['channel'] = options['VERBOSE_REPORT_CHANNEL']
    sm.send_mess(mess)


def slack_noti(mode, options, text='', title='Execute', mess=None):
    if options['BOT_TOKEN'] == "None":
        return

    sm = Messages(options)
    if not mess:
        mess = {
            'author_name': options['LOCAL_NAME'],
            'title': title,
            'content': text,
        }

    if mode == 'status':
        sm.send_info(mess)
    elif mode == 'good':
        sm.send_good(mess)
    elif mode == 'log':
        sm.send_log(mess)


def slack_file(mode, options, filename='', title='Done', mess=None):
    if options['BOT_TOKEN'] == "None":
        return
    try:
        sm = Messages(options)
        if not mess:
            mess = {
                'filename': filename,
                'title': title
            }

        if mode == 'std':
            mess['channel'] = options['STDS_CHANNEL']
        elif mode == 'verbose-report':
            mess['channel'] = options['VERBOSE_REPORT_CHANNEL']
        elif mode == 'report':
            mess['channel'] = options['REPORT_CHANNEL']

        sm.send_file(mess)
    except:
        print("Fail to send {0} to slack".format(mess['filename']))


class Messages():
    """Slack Messages sending"""

    def __init__(self, slack_options):
        self.slack_options = slack_options

    def testing(self):
        mess = {
            'author_name': 'JJ',
            'title': 'Testing title',
            'content': 'Wow ww',
        }
        self.send_info(mess)

    def send_mess(self, mess):
        self.send_attachment(mess)

    def send_info(self, mess):
        mess['color'] = '#005b9f'
        mess['icon'] = get_emoji()
        self.send_attachment(mess)

    def send_log(self, mess):
        mess['color'] = '#8BE9FD'
        mess['channel'] = self.slack_options['LOG_CHANNEL']
        mess['icon'] = get_emoji()
        self.send_attachment(mess)

    def send_good(self, mess):
        mess['color'] = '#32cb00'
        mess['icon'] = get_emoji()
        self.send_attachment(mess)

    def send_bad(self, mess):
        mess['color'] = '#c62828'
        mess['icon'] = 'https://emoji.slack-edge.com/TC2BSM362/jemoji3/9d4b76412c4c1e1a.png'
        self.send_attachment(mess)

    def send_file(self, mess):
        try:
            with open(mess['filename'], 'r+') as f:
                mess['content'] = f.read()
            mess['title'] = get_value(mess, 'title', mess['filename'])
            self.send_snippet(mess)
        except:
            with open(mess['filename'], 'r+') as f:
                mess['content'] = f.read().encode('utf-8')
            mess['title'] = get_value(mess, 'title', mess['filename'])
            self.send_snippet(mess)

    ######
    ### base sending

    def send_attachment(self, mess):
        url = 'https://slack.com/api/chat.postMessage'
        #config stuff
        token = self.slack_options['BOT_TOKEN']
        name = self.slack_options['LOCAL_NAME']

        #message stuff
        channel = get_value(
            mess, 'channel', self.slack_options['STATUS_CHANNEL'])

        author_name = get_value(mess, 'author_name')

        title = mess['title']
        content = get_value(mess, 'content', '')
        image_url = get_value(mess, 'image_url', '')
        color = mess['color']
        icon = mess['icon']

        #message body
        json_body = {
            "channel": channel,
            "attachments": [
                {
                    "image_url": image_url,
                    "author_name": author_name,
                    "fallback": title,
                    "title": title,
                    "text": content,
                    "footer": name,
                    "color": color,
                    "footer_icon": icon,
                    "ts": get_ts()
                }
            ]
        }

        r = send_JSON(url, token, json_body)

    def send_snippet(self, mess):
        token = self.slack_options['BOT_TOKEN']
        #message stuff
        channel = get_value(
            mess, 'channel', self.slack_options['REPORT_CHANNEL'])
        title = mess['title']
        filename = mess['filename']
        style = get_value(mess, 'style', 'plaintext')
        comment = get_value(mess, 'comment', '')
        content = mess['content']

        url = "https://slack.com:443/api/files.upload?channels={0}&title={1}&filename={2}&filetype={3}&initial_comment={4}&pretty=1".format(
            channel, title, filename, style, comment)

        r = send_form_boundary(url, token, content)


'''
Utilities stuff 
'''


def get_emoji():
    emojis = [
        'https://emoji.slack-edge.com/TC2BSM362/jemoji1/164b2f1f9acbaeba.png',
        'https://emoji.slack-edge.com/TC2BSM362/jemoji2/d759e0cca7025869.png',
        'https://emoji.slack-edge.com/TE7PBEN9H/jemoji1/8774d419e10124c2.png',
        'https://emoji.slack-edge.com/TE7PBEN9H/jemoji2/2b6950cdfaad8087.png',
        'https://emoji.slack-edge.com/TE7PBEN9H/jemoji3/62f09c4a09e334d8.png',
        'https://emoji.slack-edge.com/TE7PBEN9H/jemoji4/fa10edce4f481653.png',
        'https://emoji.slack-edge.com/TE7PBEN9H/jemoji5/dd199bf8963d863c.png'
        # ''
    ]
    return random.choice(emojis)


def get_ts():
    return int(time.time())


def get_value(mess, value, default=''):
    try:
        return mess[value]
    except:
        return default


'''
Sending HTTP stuff
'''


def send_GET(url, token):
    headers = {"Authorization": "Bearer " + token, "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
               "Accept-Language": "en-US,en;q=0.5", "Accept-Encoding": "gzip, deflate", "Connection": "close"}

    r = requests.get(url, headers=headers)
    return r


def send_POST(url, token, json):
    url = "https://slack.com:443/api/chat.postMessage"
    headers = {"Accept": "*/*", "Authorization": "Bearer " + token,
               "Content-type": "application/json", "Connection": "close"}

    requests.post(url, headers=headers, json=json)


def send_JSON(url, token, json_body):
    headers = {"Accept": "*/*", "Authorization": "Bearer " + token,
               "Content-type": "application/json", "Connection": "close"}
    r = requests.post(url, headers=headers, json=json_body)
    return r


def send_form_boundary(url, token, content):
    headers = {"Authorization": "Bearer " + token, "Connection": "close", "Accept": "*/*", "Origin": "https://api.slack.com", "User-Agent": "C2S v1.0",
               "Content-Type": "multipart/form-data; boundary=----WebKitFormBoundary15cPP1sBX8XrBrjn", "Accept-Encoding": "gzip, deflate", "Accept-Language": "en-US,en;q=0.9,vi;q=0.8"}

    data = "------WebKitFormBoundary15cPP1sBX8XrBrjn\r\nContent-Disposition: form-data; name=\"content\"\r\n\r\n" + \
        content + "\r\n------WebKitFormBoundary15cPP1sBX8XrBrjn--\r\n"

    r = requests.post(url, headers=headers, data=data)
    return r
