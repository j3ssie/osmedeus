import os
import json
import sys
import random
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))
import telegram
from lib.core import utils
from emoji import emojize

'''
Sending message to telegram
'''
# shortcut for sending noti to telegram
def telegram_notification(noti_type, options, output=None):
    if not options.get('TELEGRAM'):
        return
    if output and 'report' in noti_type:
        if type(output) == list:
            output = list(set(output))
            for report in output:
                if utils.not_empty_file(report):
                    telegram_file(options, report)
        else:
            if utils.not_empty_file(output):
                telegram_file(options, output)
    else:
        telegram_noti(options, str(noti_type))


def telegram_noti(options, noti_type):
    try:
        if 'log' in noti_type:
            pass
        elif 'done' in noti_type:
            telegram_done(options)
        else:
            telegram_status(options)
    except:
        pass


def telegram_status(options):
    channel = options.get('TELEGRAM_STATUS_CHANNEL', None)
    if channel is None:
        return
    module = options.get('CURRENT_MODULE')

    target = options.get('TARGET')
    message = f'#{target} - :ghost: Start *{module}* on *{target}*'

    client = telegram.Bot(token=options.get('TELEGRAM_BOT_TOKEN'))
    client.send_message(
        chat_id=channel,
        text=emojize(message, use_aliases=True),
        parse_mode=telegram.ParseMode.MARKDOWN
    )


def telegram_done(options):
    channel = options.get('TELEGRAM_STATUS_CHANNEL')
    module = options.get('CURRENT_MODULE')
    target = options.get('TARGET')
    emoji = get_emoji()
    message = f'#{target} - {emoji} Done *{module}* on *{target}*'
    client = telegram.Bot(token=options.get('TELEGRAM_BOT_TOKEN'))
    client.send_message(
        chat_id=channel,
        text=emojize(message, use_aliases=True),
        parse_mode=telegram.ParseMode.MARKDOWN
    )


def telegram_file(options, filename, token=None):
    try:
        if not token:
            token = options.get('TELEGRAM_BOT_TOKEN')
        client = telegram.Bot(token=token)
        channel = options.get('TELEGRAM_REPORT_CHANNEL')
        if channel is None:
            return
        
        emoji = get_emoji()

        target = options.get('TARGET')
        module = options.get('CURRENT_MODULE')
        message = f'#{target} - {emoji} Report *{module}* on *{target}*'

        client.send_document(
            chat_id=channel,
            document=open(filename, 'rb'),
            caption=emojize(message, use_aliases=True),
            parse_mode=telegram.ParseMode.MARKDOWN
        )
    except:
        pass


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
