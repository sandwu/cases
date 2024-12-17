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