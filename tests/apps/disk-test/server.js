const express = require('express');
const fs = require('fs');
const path = require('path');
const { exec } = require('child_process');
const { v4: uuidv4 } = require('uuid');

const app = express();
const PORT = process.env.PORT || 3000;
const DATA_DIR = '/data';

// Asegurar que la carpeta de trabajo existe
if (!fs.existsSync(DATA_DIR)) {
    fs.mkdirSync(DATA_DIR);
}

// Uso real del filesystem que contiene /data (porcentaje usado)
app.get('/disk-usage', (_req, res) => {
    exec('df -k /data', (err, stdout) => {
        if (err) return res.status(500).send('Error al obtener uso de disco');
        const lines = stdout.trim().split('\n');
        if (lines.length < 2) return res.status(500).send('Salida df inesperada');
        const parts = lines[1].trim().split(/\s+/);
        const sizeKB = parseInt(parts[1], 10);
        const usedKB = parseInt(parts[2], 10);
        const availKB = parseInt(parts[3], 10);
        const usage = (usedKB / sizeKB) * 100;
        res.json({
            usage: Number(usage.toFixed(2)),
            sizeBytes: sizeKB * 1024,
            usedBytes: usedKB * 1024,
            availBytes: availKB * 1024
        });
    });
});

// Escribir en disco (archivos únicos)
app.post('/write', (req, res) => {
    const filename = `file_${uuidv4()}.dat`;
    const filepath = path.join(DATA_DIR, filename);
    const data = 'A'.repeat(1 * 1024 * 1024); // 1MB

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