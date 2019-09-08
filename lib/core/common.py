from rest_framework.response import Response

'''
Desribe some status code here
'''


def message(status_code=200, msg="Nothing"):
    content = {'status': status_code, 'message': msg}
    return Response(content)


def returnJSON(content, status_code=200):
    content['status'] = status_code
    return Response(content)
