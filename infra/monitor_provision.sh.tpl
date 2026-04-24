export MONITOR_PUB_KEY="${monitor_pub_key}"  # Template variable is lowercase

set -euo pipefail

if [ -z "$${MONITOR_PUB_KEY}" ]; then
    echo "ERROR: SSH_PUB_KEY is not set. Please set it in your host environment."
    exit 1
fi

if ! grep -qF "$MONITOR_PUB_KEY" /root/.ssh/authorized_keys; then
    echo "$MONITOR_PUB_KEY" >> /root/.ssh/authorized_keys
fi