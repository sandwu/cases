bool CTrayWndProxy::IsProcessNeedFilter(DWORD dwPId)
{
        WTS_PROCESS_INFOW* pProcessInfo = nullptr;
        DWORD dwCount = 0 ;
        if(!WTSEnumerateProcessesW(WTS_CURRENT_SERVER_HANDLE, 0, 1, &pProcessInfo, &dwCount)) {
                LOG_WARN("WTSEnumerateProcesses failed %u",GetLastError());
                return false;
        }

        DWORD dwIndex = 0 ;
        for(dwIndex = 0 ; dwIndex < dwCount ; dwIndex++) {
                if (pProcessInfo[dwIndex].ProcessId == dwPId) {// ڵ ǰ  ϵͳ     б   
                        break;
                }        
        }

        if (dwIndex >= dwCount) {
                WTSFreeMemory(pProcessInfo);
                return true;
        }
        
        bool bIsInSystemProcessLst = false;
        for (auto& item : m_sysProcFilter) {
                if (_wcsnicmp(item.wszProcessName, 
                        pProcessInfo[dwIndex].pProcessName, 
                        item.dwProcessNameLen) == 0) {
                        LOG_INFO("process(%d) %S is systemtray so filtered");
                        bIsInSystemProcessLst = true;
                        break;
                }
        }

        WTSFreeMemory(pProcessInfo);
        return bIsInSystemProcessLst;
}


void CAppProcManager::Run()
{
        DWORD nRet = 0;

        /*   ؽ  ̵Ľ    */ 
        while(m_bRunning) {
                if(!m_bMonitoring) {
                        continue; 
                }
                /* ÿ0.5    һ   */
                Sleep(500);

                /*    xRemoteAppInit32.exe */
                MonitorXInit32(m_hMainWnd);

                std::unique_lock<std::mutex> lock(m_locker);

                auto it = m_AppMap.begin();
                while(it != m_AppMap.end()) {
                        HANDLE hProcess = it->second.hProcess;
                        nRet = WaitForSingleObject(hProcess, 0);        /*     һ ¼    */ 
                        if (nRet == WAIT_TIMEOUT) {
                                ++it;
                                continue;
                        }
                        /*      ˳    ɾ         ͽ  ̽     ֪ͨ  Ϣ*/ 
                        PostMessageToAppShel(WM_PROCESS_END, it->second.dwAppProcId, 0);
                        LOG_INFO("Erase Processpid %d from Monitor.After erase the size is %d",
                                it->second.dwAppProcId, m_AppMap.size() - 1);
                        if (it->second.hProcess) {
                                CloseHandle(it->second.hProcess);
                        }
                        m_AppMap.erase(it++);        /*   ȫ  ɾ  map   ƶ ָ 뷽   */
                }
        }
}