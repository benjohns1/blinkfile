FROM cypress/base:18.6.0
WORKDIR /cypress
COPY cypress/package.json cypress/package-lock.json ./
RUN npm ci
COPY cypress ./
ARG CYPRESS_BASE_URL=http://blinkfile:8000
ENV CYPRESS_BASE_URL=$CYPRESS_BASE_URL
ENTRYPOINT ["npm", "run", "test"]
