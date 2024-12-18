def check_application_key(func):
    """application_key合法性"""

    @wraps(func)
    def inner(*args, **kwargs):
        application_key = request.headers.get('api-key')
        if not application_key:
            raise AuthFailError('header lack application_key')

        application = ApplicationService.dao.get_by_application_key(application_key)
        if not application:
            raise AuthFailError('application_key invalid')
        request.json['application_name'] = application.application_name
        request.json['application_key'] = application.application_key
        return func(*args, **kwargs)

    return inner