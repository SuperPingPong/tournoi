FROM python:3.10.9

RUN useradd -m -d /app -s /bin/bash worker

USER worker
WORKDIR /app

ENV PATH="/app/.local/bin:${PATH}"
RUN pip install --user --upgrade pip

COPY --chown=worker:worker requirements.txt requirements.txt
RUN pip install --user -r requirements.txt

COPY --chown=worker:worker . .
CMD python main.py