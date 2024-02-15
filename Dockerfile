<<<<<<< Updated upstream
# The most minimalistic dockerfile possible.
=======
# The most minimumalistic dockerfile possible.
#  No embedded python support, no unit-testing, no aliases.
# ARG IMAGE=intersystemsdc/irishealth-community
>>>>>>> Stashed changes
ARG IMAGE=intersystemsdc/iris-community
FROM $IMAGE

WORKDIR /home/irisowner/dev
COPY .iris_init /home/irisowner/.iris_init

RUN --mount=type=bind,src=.,dst=. \
    iris start IRIS && \
	iris session IRIS < iris.script && \
    iris stop IRIS quietly
