/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */

/* 
 * File:   main.cpp
 * Author: vgerman
 *
 * Created on February 6, 2019, 11:34 AM
 */

#include <cstdlib>
#include <sg_client.h>
#include <cstring>
#include <iostream>

using namespace std;
typedef GoInt sc_handle_t;
sc_handle_t create_sc_client(
        const char* bind_ip,
        unsigned short* bind_port,
        int* error_code,
        const char* sc_ip,
        unsigned short sc_port,
        const char* tgt_ip,
        unsigned short tgt_port,
        const char* login,
        const char* pwd,
        long reconnectTimeout) {

    GoString go_bind_ip = {bind_ip};
    ushort go_bind_port = *bind_port;
    int go_error_code = 0;
    GoString go_sc_ip = {sc_ip};

    GoUint16 go_sc_port = sc_port;
    GoString go_tgt_ip = {tgt_ip};
    GoUint16 go_tgt_port = tgt_port;

    GoString go_login = {login};
    GoString go_pwd = {pwd};
    GoInt32 go_reconnectTimeout = reconnectTimeout;
    sc_handle_t sc_handle = create_sc_client(go_bind_ip, &go_bind_port, &go_error_code, go_sc_ip, go_sc_port, go_tgt_ip, go_tgt_port, go_login, go_pwd, go_reconnectTimeout);
    *bind_port = go_bind_port;
    *error_code = go_error_code;

    return sc_handle;
}

void add_tcp_server(sc_handle_t sc_handle, const char* bind_ip, unsigned short* bind_port, const char* tgt_ip, unsigned short tgt_port) {
    
    GoString go_bind_ip = {bind_ip};
    GoUint16 go_bind_port = 0;
    GoString go_tgt_ip = {tgt_ip};
    GoUint16 go_tgt_port = tgt_port;

    add_tcp_server(sc_handle, go_bind_ip, &go_bind_port, go_tgt_ip, go_tgt_port);
    *bind_port = go_bind_port;
}

int check_sc_error(sc_handle_t sc_handle, char* error_text, int size) {
    GoString go_error_text;

    int err = check_sc_error(sc_handle, &go_error_text);
    
    strncmp(error_text, go_error_text.p, size - 1);
    error_text[size - 1 ] = 0;
    
    return err;
}

void pause_e()
{
    std::cout << "Running. Type 'e' <enter> for exit" << std::endl;
    std::string line;
    while (std::getline(std::cin, line))
        if (line == "e" || line == "E")
            break;
}

/*
 * 
 */
#define LOCALHOST       "127.0.0.1"
int main(int argc, char** argv) {
    
    std::string httpsServerHost = LOCALHOST;
    unsigned short httpsPort = 7002;

    unsigned short tcpLocalPort = 7001;
    unsigned short tcpRemotePort = 7004;

    long reconnectTimeout = 100;
    int err = 0;
    sc_handle_t sc_handle = create_sc_client(
            LOCALHOST,
            
            &tcpLocalPort,
            &err,
            
            httpsServerHost.c_str(),
            httpsPort,

            LOCALHOST,
            tcpRemotePort,
            "",
            "",
            reconnectTimeout);

    if (err != 0) {
        std::cerr <<"Failed to connect to "<< httpsServerHost<< ":" <<httpsPort<< " ";
        return err;
    }
    
    std::cout << "Connected to  remote server " << httpsServerHost << ":" <<httpsPort<< " ";
    pause_e();
    
    delete_sc_client(sc_handle);
    return 0;
}

