import * as http from 'http';
import * as fs from 'fs';
import * as path from 'path';

const PORT = 8080;
const STATIC_DIR = path.join(__dirname, '../pages');
const REVEAL_DIR = path.join(__dirname, '../node_modules/reveal.js/dist');

const mimeTypes: Record<string, string> = {
	'.html': 'text/html',
	'.js': 'text/javascript',
	'.css': 'text/css',
	'.json': 'application/json',
	'.png': 'image/png',
	'.jpg': 'image/jpg',
	'.gif': 'image/gif',
	'.svg': 'image/svg+xml',
	'.woff': 'font/woff',
	'.woff2': 'font/woff2',
};

const server = http.createServer((req, res) => {
	if (req.url?.startsWith('/reveal.js/')) {
		let filePath = path.join(REVEAL_DIR, req.url.replace('/reveal.js/', ''));
		const ext = path.extname(filePath);
		const contentType = mimeTypes[ext] || 'application/octet-stream';
		fs.readFile(filePath, (err, content) => {
			if (err) {
				res.writeHead(404);
				res.end(`Not Found: ${req.url}`);
			} else {
				res.writeHead(200, { 'Content-Type': contentType });
				res.end(content, 'utf-8');
			}
		});
		return;
	}

	let filePath = path.join(STATIC_DIR, req.url === '/' ? 'index.html' : req.url);

	const ext = path.extname(filePath);
	const contentType = mimeTypes[ext] || 'application/octet-stream';

	fs.readFile(filePath, (err, content) => {
		if (err) {
			if (err.code === 'ENOENT') {
				fs.readFile(path.join(STATIC_DIR, '404.html'), (err, content) => {
					res.writeHead(404, { 'Content-Type': 'text/html' });
					res.end(content, 'utf-8');
				});
			} else {
				res.writeHead(500);
				res.end(`Server Error: ${err.code}`);
			}
		} else {
			res.writeHead(200, { 'Content-Type': contentType });
			res.end(content, 'utf-8');
		}
	});
});

server.listen(PORT, () => {
	console.log(`Server running at http://localhost:${PORT}/`);
});
