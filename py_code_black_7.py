def sync_testbed(task_id: int):
    with Session(engine) as session:
        task = session.get(TestTask, task_id)

        testbed = task.test_bed.variable_config
        testbed_dict = yaml.safe_load(testbed)

        links = task.test_bed.agent_keyword_service_links
        for link in links:
            host = link.agent.host
            flag, msg = send_testbed(testbed_dict, task_id, host)
            if not flag:
                task_commit(task, session, "失败", msg)
                return False
        return True