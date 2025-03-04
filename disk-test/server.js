const express = require('express');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { v4: uuidv4 } = require('uuid');

const app = express();
const PORT = process.env.PORT || 3000;
const DATA_DIR = path.join(__dirname, 'data');

// Asegurar que la carpeta de trabajo existe
if (!fs.existsSync(DATA_DIR)) {
    fs.mkdirSync(DATA_DIR);
}

// Endpoint para obtener el uso del disco
app.get('/disk-usage', (req, res) => {
    fs.stat(DATA_DIR, (err, stats) => {
        if (err) return res.status(500).send('Error al obtener información del disco');

        const diskUsage = stats.blocks * stats.blksize; // Uso en bytes
        const totalSpace = os.totalmem(); // Simulación de total (en un entorno real usaría un metodo diferente)
        const usagePercentage = (diskUsage / totalSpace) * 100;

        res.json({ usage: usagePercentage.toFixed(2) });
    });
});

// Escribir en disco (archivos únicos)
app.post('/write', (req, res) => {
    const filename = `file_${uuidv4()}.dat`;
    const filepath = path.join(DATA_DIR, filename);
    const data = 'A'.repeat(10 * 1024 * 1024); // 10MB

    fs.writeFile(filepath, data, (err) => {
        if (err) return res.status(500).send('Error al escribir en disco');
        res.send(`Archivo ${filename} creado`);
    });
});

// Leer un archivo aleatorio
app.get('/read', (req, res) => {
    fs.readdir(DATA_DIR, (err, files) => {
        if (err || files.length === 0) return res.status(500).send('No hay archivos para leer');

        const randomFile = files[Math.floor(Math.random() * files.length)];
        const filepath = path.join(DATA_DIR, randomFile);

        fs.readFile(filepath, (err, data) => {
            if (err) return res.status(500).send('Error al leer el archivo');
            res.send(`Archivo leído: ${randomFile}, tamaño: ${data.length} bytes`);
        });
    });
});

// Borrar archivos más antiguos cuando sea necesario
app.delete('/delete', (req, res) => {
    fs.readdir(DATA_DIR, (err, files) => {
        if (err || files.length === 0) return res.send('No hay archivos para eliminar');

        // Ordenar archivos por fecha de modificación (más antiguos primero)
        files.sort((a, b) => {
            return fs.statSync(path.join(DATA_DIR, a)).mtimeMs -
                fs.statSync(path.join(DATA_DIR, b)).mtimeMs;
        });

        // Borrar el archivo más antiguo
        const oldestFile = path.join(DATA_DIR, files[0]);
        fs.unlink(oldestFile, (err) => {
            if (err) return res.status(500).send('Error al borrar archivo');
            res.send(`Archivo eliminado: ${files[0]}`);
        });
    });
});

// Iniciar el servidor
app.listen(PORT, () => console.log(`Servidor corriendo en el puerto ${PORT}`));