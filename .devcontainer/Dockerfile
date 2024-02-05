FROM --platform=linux/amd64 pulumi/pulumi:3.104.2

RUN <<EOF
apt install -y bash-completion vim tree

## Set bash completion for pulumi
pulumi gen-completion bash >> /etc/bash_completion.d/pulumi

## Install azure cli
curl -sL https://aka.ms/InstallAzureCLIDeb | bash
EOF
