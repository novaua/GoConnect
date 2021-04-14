#https://www.sslshopper.com/article-most-common-openssl-commands.html

#Generate a new private key and Certificate Signing Request
#openssl req -out CSR.csr -new -newkey rsa:2048 -nodes -keyout privateKey.key

#Generate a self-signed certificate
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -keyout privateKey.key -out certificate.crt

openssl x509 -inform der -in certificate.cer -out certificate.pem
