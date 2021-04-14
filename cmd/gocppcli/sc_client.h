#ifndef SC_CLIENT_H
#define SC_CLIENT_H

#ifdef WIN32
#define SC_CLIENT_PUBLIC_API __declspec(dllexport)
#else
#define SC_CLIENT_PUBLIC_API
#endif

#ifdef WIN32
#define API_CTP __cdecl
#else
#define API_CTP
#endif

#ifdef __cplusplus
extern "C" {
#endif

SC_CLIENT_PUBLIC_API void*
    API_CTP
create_sc_client(const char* bind_ip,
                 unsigned short* bind_port,
                 int* error_code,
                 const char* sc_ip,
                 unsigned short sc_port,
                 const char* tgt_ip,
                 unsigned short tgt_port,
                 const char* login,
                 const char* pwd,
                 long reconnectTimeout);

SC_CLIENT_PUBLIC_API void
    API_CTP
add_tcp_server(void* sc, const char* bind_ip, unsigned short* bind_port, const char* tgt_ip, unsigned short tgt_port);

SC_CLIENT_PUBLIC_API void
    API_CTP
delete_sc_client(void* sc);

SC_CLIENT_PUBLIC_API int
    API_CTP
check_sc_error(void* sc, char* error_text, int size);

#ifdef __cplusplus
}
#endif

#endif //SC_CLIENT_H
