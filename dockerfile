FROM python:3.7-slim

ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Update package lists and install required packages, including nginx
RUN apt-get update && \
    apt-get install -y git curl ca-certificates libcurl4-openssl-dev vim unzip && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /opt/ScopeSentry/

# Copy ScopeSentry project files to the working directory
COPY dist/ScopeSentry_linux_amd64_v1/ScopeSentry /opt/ScopeSentry/

RUN chmod +x /opt/ScopeSentry/ScopeSentry

ENTRYPOINT ["/opt/ScopeSentry/ScopeSentry"]

# # Copy the start.sh script into the container
# RUN cp /opt/ScopeSentry/start.sh /usr/local/bin/start.sh
#
# # Grant execute permission to the start.sh script
# RUN chmod +x /usr/local/bin/start.sh
#
# # Start the container using start.sh
# CMD ["/usr/local/bin/start.sh"]