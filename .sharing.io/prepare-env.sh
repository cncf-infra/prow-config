#!/usr/bin/env bash

GIT_ROOT=$(git rev-parse --show-toplevel)

echo "===="
echo "Prow"
echo "===="
echo
echo "Please fork this repo to continue."
read -r -p "Press enter to continue, or C-c to cancel"
echo
echo "# Env set up"
echo "TODO: Navigate to 'https://github.com/settings/developers' -> OAuth Apps"
echo "      go to an existing or new OAuth app."
echo "      ensure that:"
echo "        - homepage URL is set to 'https://prow.${SHARINGIO_PAIR_BASE_DNS_NAME}'"
echo "        - authorization callback URL is set to 'https://prow.${SHARINGIO_PAIR_BASE_DNS_NAME}/oauth'"
echo
if [ ! -f $GIT_ROOT/.sharing.io/.oauth-env ]; then
    echo "Input:"
    read -r -p "OAUTH_CLIENT_ID (github oauth app client id)                  : " OAUTH_CLIENT_ID
    read -r -p "OAUTH_CLIENT_SECRET (github oauth app client generated secret): " OAUTH_CLIENT_SECRET
    cat <<EOF > $GIT_ROOT/.sharing.io/.oauth-env
OAUTH_CLIENT_ID=${OAUTH_CLIENT_ID}
OAUTH_CLIENT_SECRET=${OAUTH_CLIENT_SECRET}
EOF
else
    echo "There already appears to be OAuth env set, to edit: checkout '$GIT_ROOT/.sharing.io/.oauth-env'."
    read -r -p "Press enter to continue to exit this prompt"
fi
touch $GIT_ROOT/.sharing.io/setup-complete

