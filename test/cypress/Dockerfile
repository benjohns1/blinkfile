FROM cypress/base:18.6.0
WORKDIR /cypress
COPY package.json package-lock.json ./
RUN npm ci
COPY . ./
ENTRYPOINT ["npm", "run", "test"]
