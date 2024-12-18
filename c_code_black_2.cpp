cellularif_auth_protocol_t cellularif_auth_protocol_from_str(const char* str)
{
    if (str == NULL)
        return CELLULARIF_AUTH_PROTOCOL_NONE;

    if (strncmp(str, "AUTO", strlen(str)) == 0)
        return CELLULARIF_AUTH_PROTOCOL_AUTO;
    if (strncmp(str, "PAP", strlen(str)) == 0)
        return CELLULARIF_AUTH_PROTOCOL_PAP;
    if (strncmp(str, "CHAP", strlen(str)) == 0)
        return CELLULARIF_AUTH_PROTOCOL_CHAP;

    return CELLULARIF_AUTH_PROTOCOL_NONE;
}