# client_id is *only* for the output filenames
# incrementing the serial number is important
CLIENT_ID="client"
CLIENT_SERIAL=01

###### PICK ONE OF THE TWO FOLLOWING ######
###### (instrux in the CA section above) ######
# rsa
openssl genrsa -aes256 -passout pass:xxxx -out ${CLIENT_ID}.pass.key 4096
openssl rsa -passin pass:xxxx -in ${CLIENT_ID}.pass.key -out ${CLIENT_ID}.key
rm ${CLIENT_ID}.pass.key
# ec
openssl ecparam -genkey -name secp256r1 | openssl ec -out ${CLIENT_ID}.key
###### END  "PICK ONE" SECTION ######

# whichever you picked, you should now have a `client.key` file.

# generate the CSR
# i think the Common Name is the only important thing here. think of it like
# a display name or login.
openssl req -new -key ${CLIENT_ID}.key -out ${CLIENT_ID}.csr

# issue this certificate, signed by the CA root we made in the previous section
openssl x509 -req -days 3650 -in ${CLIENT_ID}.csr -CA ca.pem -CAkey ca.key -set_serial ${CLIENT_SERIAL} -out ${CLIENT_ID}.pem