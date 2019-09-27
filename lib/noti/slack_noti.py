import os
import json
import sys
import random
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))
import slack
from lib.core import utils

'''
Sending message to Slack
'''


def slack_monitor(options, filename, monitor_type='new'):
    if options.get('SLACK_MONITOR_TOKEN') and options.get('SLACK_MONITOR_TOKEN') != 'None':
        token = options.get('SLACK_MONITOR_TOKEN')
    else:
        token = options.get('SLACK_BOT_TOKEN')

    client = slack.WebClient(token=token)

    if monitor_type == 'new':
        channel = options.get('NEW_CHANNEL', None)
        if channel is None:
            return
    elif monitor_type == 'missing':
        channel = options.get('MISSING_CHANNEL', None)
        if channel is None:
            return

    client.files_upload(
        channels=channel,
        file=filename,
        title=filename,
        filetype='text'
    )


# shortcut for sending noti to slack
def slack_notification(noti_type, options, output=None):
    utils.print_info(f"Sending {noti_type} notification to slack")
    if not options.get('SLACK'):
        return

    if output and 'report' in noti_type:
        if type(output) == list:
            output = list(set(output))
            for report in output:
                if utils.not_empty_file(report):
                    slack_file(options, report)
        else:
            if utils.not_empty_file(output):
                slack_file(options, output)
    else:
        slack_noti(options, str(noti_type))


def slack_noti(options, noti_type):
    try:
        if 'log' in noti_type:
            pass
        elif 'done' in noti_type:
            slack_done(options)
        else:
            slack_status(options)
    except:
        utils.print_bad("Fail to send noti to slack")
        pass


def slack_status(options):
    channel = options.get('STATUS_CHANNEL', None)
    if channel is None:
        return
    module = options.get('CURRENT_MODULE')

    target = options.get('TARGET')
    message = f':ghost: Start *{module}* on *{target}*'
    
    client = slack.WebClient(token=options.get('SLACK_BOT_TOKEN'))
    client.chat_postMessage(
        channel=channel,
        blocks=[
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": message,
                },
            },
            {
                "type": "divider"
            }
        ]
    )


def slack_done(options):
    channel = options.get('STATUS_CHANNEL')
    module = options.get('CURRENT_MODULE')
    target = options.get('TARGET')
    emoji = get_emoji()
    message = f'{emoji} Done *{module}* on *{target}*'
    client = slack.WebClient(token=options.get('SLACK_BOT_TOKEN'))
    client.chat_postMessage(
        channel=channel,
        blocks=[
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": message,
                },
                # "accessory": {
                #     "type": "image",
                #     "image_url": get_emoji(),
                #     "alt_text": 'status',
                # }
            },
            {
                "type": "divider"
            }
        ]
    )


def slack_file(options, filename, token=None):
    if not token:
        token = options.get('SLACK_BOT_TOKEN')
    client = slack.WebClient(token=token)
    channel = options.get('REPORT_CHANNEL')
    if channel is None:
        return

    client.files_upload(
        channels=channel,
        file=filename,
        title=filename,
        filetype='text'
    )


def get_emoji():
    emojis = [
        ':robot_face:',
        ':alien:',
        ':gift:',
        ':gun:',
        ':diamond_shape_with_a_dot_inside:',
        ':rocket:',
        ':bug:',
        ':broccoli:',
        ':shamrock:',
    ]
    return random.choice(emojis)
