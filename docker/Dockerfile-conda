FROM ubuntu:22.04

ARG MINIFORGE_ARCH="x86_64"
ARG MINIFORGE_VERSION=22.9.0-3

ENV CONDA_DIR /opt/conda
ENV PATH "${CONDA_DIR}/bin:${PATH}"
ENV WP_UID 1000
ENV WP_USER workspace
ENV HOME /home/workspace

RUN useradd -M -s /bin/bash -N -u ${WP_UID} ${WP_USER} \
 && mkdir -p ${HOME} \
 && chown -R ${WP_UID}:users ${HOME} \
 && chown -R ${WP_UID}:users /usr/local/bin 

RUN export DEBIAN_FRONTEND=noninteractive \
 && apt-get -yq update \
 && apt-get -yq install --no-install-recommends \
    apt-transport-https \
    bash \
    bzip2 \
    ca-certificates \
    curl \
    git \
    gnupg \
    gnupg2 \
    locales \
    lsb-release \
    nano \
    software-properties-common \
    tzdata \
    unzip \
    vim \
    wget \
    zip \
    xz-utils \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

RUN mkdir -p ${CONDA_DIR} \
 && chown -R ${WP_UID}:users ${CONDA_DIR}

RUN curl -sL "https://github.com/conda-forge/miniforge/releases/download/${MINIFORGE_VERSION}/Miniforge3-${MINIFORGE_VERSION}-Linux-${MINIFORGE_ARCH}.sh" -o /tmp/Miniforge3.sh \
 && curl -sL "https://github.com/conda-forge/miniforge/releases/download/${MINIFORGE_VERSION}/Miniforge3-${MINIFORGE_VERSION}-Linux-${MINIFORGE_ARCH}.sh.sha256" -o /tmp/Miniforge3.sh.sha256 \
 && echo "$(cat /tmp/Miniforge3.sh.sha256 | awk '{ print $1; }') /tmp/Miniforge3.sh" | sha256sum --check \
 && rm /tmp/Miniforge3.sh.sha256 \
 && /bin/bash /tmp/Miniforge3.sh -b -f -p ${CONDA_DIR} \
 && rm /tmp/Miniforge3.sh \
 && conda config --system --set auto_update_conda false \
 && conda config --system --set show_channel_urls true \
 && echo "conda ${MINIFORGE_VERSION}" >> ${CONDA_DIR}/conda-meta/pinned \
 && conda update -y -q --all \
 && conda clean -a -f -y

USER ${WP_UID}

WORKDIR ${HOME}