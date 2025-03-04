# Usar la imagen base de Node.js
FROM node:18

# Establecer el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar package.json y package-lock.json
COPY package*.json ./

# Instalar las dependencias
RUN npm install

# Copiar el resto de los archivos de la aplicación
COPY . .

# Exponer el puerto 3000 (o el que hayas definido en tu código)
EXPOSE 3000

# Comando para ejecutar la aplicación
CMD ["node", "server.js"]