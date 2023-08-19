# Generate test localhost certificates

main:
	@ make root
	@ make localhost
	@ make view-root
	@ make view-localhost

root:
	@ openssl req -x509 -nodes -new \
		-sha256 \
		-days 1024 \
		-newkey rsa:2048 \
		-keyout root_ca.key \
		-out root_ca.pem \
		-subj "/C=US/CN=Spec-Root-CA"
	@ openssl x509 \
		-outform pem \
		-in root_ca.pem \
		-out root_ca.crt

localhost:
	@ openssl req -new -nodes \
		-newkey rsa:2048 \
		-keyout localhost.key \
		-out localhost.csr \
		-subj "/C=US/ST=NoState/L=NoCity/O=Spec-Certificates/CN=localhost"
	@ openssl x509 \
		-req -sha256 \
		-days 1024 \
		-in localhost.csr \
		-CA root_ca.pem \
		-CAkey root_ca.key \
		-CAcreateserial \
		-extfile localhost.ext \
		-out localhost.crt

view-root:
	@ openssl x509 -in root_ca.crt -text -noout
view-localhost:
	@ openssl x509 -in localhost.crt -text -noout
