import os
import sys
import requests
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
sys.path.append(os.path.dirname(os.path.dirname(os.path.realpath(__file__))))

from lib.core import utils

osmedeus_headers = {"User-Agent": "Osmedeus/v2.0", "Accept": "*/*",
           "Content-type": "application/json", "Connection": "close"}

# send request through Burp proxy for debug purpose
PROXY = {
    'http': 'http://127.0.0.1:8081',
    'https': 'http://127.0.0.1:8081'
}


# just for sending cmd purpose
def post_without_response(url, body, headers=osmedeus_headers):
    try:
        r = requests.post(url, verify=False, headers=headers,
                          json=body, timeout=2)
        return r
    except:
        return False


def send_get(url, data, headers=osmedeus_headers,
             cookies=None, is_json=False, proxy=None, retries=3):

    for times in range(retries):
        try:
            r = just_send_get(url, data, headers=headers,
                              cookies=cookies, is_json=is_json, proxy=proxy)
            return r
        except Exception:
            # retry if something went wrong
            utils.random_sleep(3, 6)
            # return nothing after few times retry
            if times == int(retries - 1):
                return None


def send_post(url, data, headers=osmedeus_headers,
              cookies=None, is_json=False, proxy=None, retries=3):

    for times in range(retries):
        try:
            r = just_send_post(url, data, headers=headers,
                               cookies=cookies, is_json=is_json, proxy=proxy)
            return r
        except Exception:
            # retry if something went wrong
            utils.random_sleep(3, 6)
            # return nothing after few times retry
            if times == int(retries - 1):
                return None


# really send GET request
def just_send_get(url, data=None, headers=osmedeus_headers,
                  cookies=None, is_json=False, proxy=None):
    if proxy:
        if is_json:
            r = requests.get(url, verify=False,
                             headers=headers, cookies=cookies,
                             json=data, proxies=proxy)
        elif data:
            r = requests.get(url, verify=False,
                             headers=headers, cookies=cookies,
                             data=data, proxies=proxy)
            pass
        else:
            r = requests.get(url, verify=False, headers=headers,
                             cookies=cookies, proxies=proxy)
    else:
        if is_json:
            r = requests.get(url, verify=False,
                             headers=headers, cookies=cookies, json=data)
        elif data:
            r = requests.get(url, verify=False,
                             headers=headers, cookies=cookies,
                             data=data, proxies=proxy)
        else:
            r = requests.get(url, verify=False,
                             headers=headers, cookies=cookies)

    return r


# really send POST request
def just_send_post(url, data, headers=osmedeus_headers,
                   cookies=None, is_json=False, proxy=None):
    if proxy:
        if is_json:
            r = requests.post(url, verify=False,
                              headers=headers, cookies=cookies,
                              json=data, proxies=proxy)
        elif data:
            r = requests.post(url, verify=False,
                              headers=headers, cookies=cookies,
                              data=data, proxies=proxy)
        else:
            r = requests.post(url, verify=False,
                              headers=headers, cookies=cookies, proxies=proxy)
    else:
        if is_json:
            r = requests.post(url, verify=False,
                              headers=headers, cookies=cookies,
                              json=data)
        elif data:
            r = requests.post(url, verify=False,
                              headers=headers, cookies=cookies,
                              data=data)
        else:
            r = requests.post(url, verify=False,
                              headers=headers, cookies=cookies)

    return r
