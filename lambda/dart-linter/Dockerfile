FROM public.ecr.aws/lambda/python:3.12

WORKDIR ${LAMBDA_TASK_ROOT}

RUN dnf install -y unzip && \
    python -c "import urllib.request; urllib.request.urlretrieve('https://storage.googleapis.com/dart-archive/channels/stable/release/latest/sdk/dartsdk-linux-x64-release.zip', '/tmp/dart.zip')" && \
    unzip -q /tmp/dart.zip -d /opt && \
    ln -s /opt/dart-sdk/bin/dart /usr/local/bin/dart && \
    dnf clean all && \
    rm -f /tmp/dart.zip

COPY index.py ./

CMD ["index.handler"]
