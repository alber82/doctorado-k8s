# Usar una imagen base de Node.js
FROM node:18

# Establecer el directorio de trabajo dentro del contenedor
WORKDIR /app

# Copiar los archivos de configuración package.json y package-lock.json
COPY tests/apps/disk-test/package*.json ./

# Instalar las dependencias
RUN npm install

# Copiar el resto del código del servidor
COPY tests/apps/disk-test/ . 

# Exponer el puerto donde el servidor escucha
EXPOSE 3000

# Comando para iniciar la app
CMD ["npm", "start"]
