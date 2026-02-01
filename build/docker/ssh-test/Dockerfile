FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    openssh-server \
    sudo \
    curl \
    wget \
    vim \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir /var/run/sshd

# Create test user with sudo access
RUN useradd -m -s /bin/bash testuser \
    && echo "testuser:testpass" | chpasswd \
    && usermod -aG sudo testuser \
    && echo "testuser ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Setup SSH directory for testuser
RUN mkdir -p /home/testuser/.ssh \
    && chmod 700 /home/testuser/.ssh

# Copy authorized keys
COPY id_ed25519.pub /home/testuser/.ssh/authorized_keys

# Set proper permissions
RUN chmod 600 /home/testuser/.ssh/authorized_keys \
    && chown -R testuser:testuser /home/testuser/.ssh

# Configure SSH
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin no/' /etc/ssh/sshd_config \
    && sed -i 's/#PubkeyAuthentication yes/PubkeyAuthentication yes/' /etc/ssh/sshd_config \
    && sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config

EXPOSE 22

CMD ["/usr/sbin/sshd", "-D"]
