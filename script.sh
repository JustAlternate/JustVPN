set -e
CLIENT_PUB_IP=$1
WAITING_TIME_UNTIL_DESTROY=$2

if [[ -z "$CLIENT_PUB_IP" || -z "$WAITING_TIME_UNTIL_DESTROY" ]]; then
    echo "Usage: $0 <client_public_ip> <waiting_time_until_destroy>"
    exit 1
fi

terraform apply -var="endpoint=$CLIENT_PUB_IP" -var-file="secrets.tfvars" --auto-approve
HOST_IP=$(terraform output instance_ip | tr -d '"')
echo "Host_IP: $HOST_IP"

sleep 30

SERVER_PUB_KEY=$(/usr/bin/env ssh -i ~/.ssh/id_ed25519 -o "StrictHostKeyChecking no" -q root@$HOST_IP -t 'cat wg-public.key')
echo "Server_Pub_KEY: $SERVER_PUB_KEY"

sleep "$WAITING_TIME_UNTIL_DESTROY"
terraform destroy -var="endpoint=$CLIENT_PUB_IP" -var-file="secrets.tfvars" --auto-approve
