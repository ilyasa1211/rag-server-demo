FROM golang:1.25-bookworm AS base

FROM base

ARG USERNAME=dev
ARG USER_UID=1000
ARG USER_GID=1000

# Add user
RUN groupadd --gid=${USER_GID} ${USERNAME} && \
  useradd --gid=${USER_GID} --uid=${USER_UID} --shell /usr/bin/bash --create-home ${USERNAME}

RUN apt-get update && \
  apt-get install -y sudo openssh-client git 

RUN echo "${USERNAME} ALL=(root) NOPASSWD:ALL" > /etc/sudoers.d/${USERNAME} && \
  chmod 440 /etc/sudoers.d/${USERNAME}

USER ${USERNAME}