FROM public.ecr.aws/lambda/python:3.12

WORKDIR ${LAMBDA_TASK_ROOT}

RUN dnf install -y gcc-c++ && \
    dnf clean all

COPY index.py ./

CMD ["index.handler"]
