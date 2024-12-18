def init_keyword_services(session, task: TestTask, links):
    agent_envs = []
    for link in links:
        host = link.agent.host
        if not host:
            task_commit(task, session, "失败", f"执行机{link.agent.name}状态异常，请查看执行机是否正常")
            return []
        
        # 开启服务
        config = link.keyword_service.config
        if link.agent.keyword_service_type == "k8s":
            release_name = config.get("helm_chart", "").split("/")[-1]
            body = {
                "helm_chart": config.get("helm_chart", ""),
                "release_name": release_name,
                "namespace": f"{release_name}-{str(uuid.uuid4())}",
                "values": config.get("values", "")
            }

        elif link.agent.keyword_service_type == "docker":
            body = {
                "image": config.get("image", ""),
                "args": config.get("args", "")
            }

        else:
            body = {
                "git_repo": config.get("git_repo", ""),
                "git_branch": config.get("git_branch", ""),
                "start_command": config.get("start_command", "")
            }

        response = requests.Session().post(f"http://{host}/start_service", json=body, verify=False)
        if response.status_code != 200 or response.json().get("code") != 0:
            task_commit(task, session, "失败", f"执行机{link.agent.name}启动服务失败, 错误信息: {response.text}")
            logging.error(f"{link.agent.name} start service failed, err: {response.text}")
            return []
        
        # 获取安装状态，当状态为failed或completed时，表示初始化agent结束
        # 超时时间设置为5分钟
        status, resp = get_service_status(link, body.get("namespace", ""), body.get("image", ""), host, task, session)
        if status == "failed":
            task_commit(task, session, "失败", f"执行机{link.agent.name}启动关键字服务失败，错误信息: {resp.get("message")}")
            return []
        if resp.get("env"):
            agent_envs.append(resp["env"])
    return agent_envs