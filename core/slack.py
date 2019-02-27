'''
Sending message to Slack
'''
import requests
import json
import time
import random


###Slack printing
def slack_info(options, text='', title='Execute', mess=None):
    sm = Messages(options)
    if not mess:
        mess = {
            'author_name': options['LOCAL_NAME'],
            'title': title,
            'content': text,
        }
    sm.send_info(mess)

###Slack printing
def send_log(options, text='', title='Execute', mess=None):
    sm = Messages(options)
    if not mess:
        mess = {
            'author_name': options['LOCAL_NAME'],
            'title': title,
            'content': text,
        }
    sm.send_log(mess)


def slack_good(options, text='', title='Done', mess=None):
    sm = Messages(options)
    if not mess:
        mess = {
            'author_name': options['LOCAL_NAME'],
            'title': title,
            'content': text,
        }
    sm.send_info(mess)


def slack_file(options, filename='', title='Done', mess=None):
    sm = Messages(options)
    if not mess:
        mess = {
            'filename': filename,
            'title': title
        }
    sm.send_file(mess)




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
        with open(mess['filename'], 'r+') as f:
            mess['content'] = f.read()

        mess['title'] = mess['filename']
        self.send_snippet(mess)

    ######
    ### base sending
    def send_attachment(self, mess):
        url = 'https://slack.com/api/chat.postMessage'
        #config stuff
        token = self.slack_options['BOT_TOKEN']
        name = self.slack_options['LOCAL_NAME']

        #message stuff
        channel = get_value(mess, 'channel', self.slack_options['STATUS_CHANNEL'])

        author_name = get_value(mess, 'author_name')
        
        title = mess['title']
        content = mess['content']
        color = mess['color']
        icon = mess['icon']

        #message body
        json_body = {
            "channel": channel,
            "attachments": [
                {
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
        channel = self.slack_options['REPORT_CHANNEL']
        title = mess['title']
        filename = mess['filename']
        style = get_value(mess, 'style', 'plaintext')
        comment = get_value(mess, 'comment', '')
        content = mess['content']

        url = "https://slack.com:443/api/files.upload?channels={0}&title={1}&filename={2}&filetype={3}&initial_comment={4}&pretty=1".format(channel, title, filename, style, comment)

        r = send_form_boundary(url, token, content)


'''
Utilities stuff 
'''

def get_emoji():
    emojis = [
        'https://emoji.slack-edge.com/TC2BSM362/jemoji1/164b2f1f9acbaeba.png',
        'https://emoji.slack-edge.com/TC2BSM362/jemoji2/d759e0cca7025869.png'
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
