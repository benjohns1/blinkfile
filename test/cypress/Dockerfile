FROM cypress/base:22.19.0
WORKDIR /cypress
COPY package.json package-lock.json ./
RUN npm ci
COPY . ./
ENTRYPOINT ["npm", "run", "test"]
