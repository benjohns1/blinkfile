FROM cypress/base:22.12.0
WORKDIR /cypress
COPY package.json package-lock.json ./
RUN npm ci
COPY . ./
ENTRYPOINT ["npm", "run", "test"]
