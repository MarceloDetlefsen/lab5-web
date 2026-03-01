package main

import "fmt"

// Estilos compartidos entre todas las páginas
const sharedCSS = `
	body { font-family: Arial; background: #f4f4f4; padding: 40px; }
	h1   { text-align: center; }
	a    { color: #ffb545; }
`

func indexTemplate(tableRows string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<title>Control de Series</title>
	<style>
		%s
		table {
			margin: auto;
			border-collapse: collapse;
			width: 90%%;
			background: white;
		}
		p {
			text-align: center;
			font-style: italic;
			color: #555;
		}
		th, td {
			border: 1px solid #000;
			padding: 10px;
			text-align: center;
		}
		th {
			background: #ffb545;
			color: white;
		}
		tr:nth-child(even) { background: #6ec8ff; }
		tr:nth-child(odd)  { background: #cae6ff; }
		.progress-container {
			width: 100%%;
			background-color: #ddd;
			border-radius: 10px;
			overflow: hidden;
		}
		.progress-bar {
			height: 20px;
			background-color: #4CAF50;
			text-align: center;
			color: white;
			line-height: 20px;
			font-size: 12px;
		}
		.add-link {
			display: block;
			text-align: center;
			margin-bottom: 20px;
			font-size: 16px;
		}
		.btn-next {
			background: #ffb545;
			color: white;
			border: none;
			padding: 6px 14px;
			border-radius: 6px;
			cursor: pointer;
			font-size: 14px;
			margin: 2px;
		}
		.btn-next:hover    { background: #e0a030; }
		.btn-next:disabled { background: #aaa; cursor: default; }
		tr.completed td {
			background: #d4edda !important;
			font-style: italic;
			color: #555;
		}
		input[type="text"], input[type="number"] {
			width: 90%%;
			padding: 4px;
			border: 1px solid #ffb545;
			border-radius: 4px;
		}
	</style>
</head>
<body>
	<h1>Control de Series</h1>
	<p>( No miro series :/ )<br>Solo puse datos de series que conozco, pero no son mis estadisticas.</p>
	<table>
		<tr>
			<th>ID</th>
			<th>Serie</th>
			<th>Episodio Actual</th>
			<th>Total de Episodios</th>
			<th>Progreso</th>
			<th>Alterar Progreso</th>
			<th>Eliminar Serie</th>
			<th>Editar Serie</th>
		</tr>
		%s
	</table>
	<br>
	<a class="add-link" href="/create">Agregar nueva serie</a>

	<script>
		async function nextEpisode(id, current, total) {
			if (current >= total) return;
			const response = await fetch("/update?id=" + id, { method: "POST" });
			if (response.ok) {
				location.reload();
			} else {
				alert("Error al actualizar el episodio");
			}
		}

		async function nextEpisodeMinus(id, current, total) {
			if (current <= 0) return;
			const response = await fetch("/update-minus?id=" + id, { method: "POST" });
			if (response.ok) {
				location.reload();
			} else {
				alert("Error al actualizar el episodio");
			}
		}

		async function deleteSerie(id) {
			if (!confirm('¿Eliminar esta serie?')) return;
			const response = await fetch("/delete?id=" + id, { method: "DELETE" });
			if (response.ok) {
				location.reload();
			} else {
				alert("Error al eliminar");
			}
		}

		function editSerie(id, name, current, total) {
			const row = document.getElementById("row-" + id);
			const tds = row.querySelectorAll("td");

			tds[1].innerHTML = '<input id="edit-name-' + id + '" type="text" value="' + name + '">';
			tds[2].innerHTML = '<input id="edit-current-' + id + '" type="number" value="' + current + '">';
			tds[3].innerHTML = '<input id="edit-total-' + id + '" type="number" value="' + total + '">';
			tds[7].innerHTML = '<button class="btn-next" onclick="saveEdit(' + id + ')">Guardar</button>';
		}

		async function saveEdit(id) {
			const name    = document.getElementById("edit-name-" + id).value;
			const current = document.getElementById("edit-current-" + id).value;
			const total   = document.getElementById("edit-total-" + id).value;

			const response = await fetch("/edit?id=" + id, {
				method: "PUT",
				headers: { "Content-Type": "application/x-www-form-urlencoded" },
				body: "series_name=" + encodeURIComponent(name) + "&current_episode=" + current + "&total_episodes=" + total
			});

			if (response.ok) {
				location.reload();
			} else {
				alert("Error al guardar");
			}
		}
	</script>
</body>
</html>
`, sharedCSS, tableRows)
}

func createFormTemplate() string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<title>Agregar Serie</title>
	<style>
		%s
		form {
			max-width: 400px;
			margin: auto;
			background: white;
			padding: 30px;
			border-radius: 8px;
			box-shadow: 0 2px 8px rgba(0,0,0,0.1);
		}
		label {
			display: block;
			margin-top: 15px;
			font-weight: bold;
		}
		input {
			width: 100%%;
			padding: 8px;
			margin-top: 5px;
			box-sizing: border-box;
			border: 1px solid #ccc;
			border-radius: 4px;
		}
		button {
			margin-top: 20px;
			width: 100%%;
			padding: 10px;
			background: #ffb545;
			color: white;
			border: none;
			border-radius: 4px;
			font-size: 16px;
			cursor: pointer;
		}
		button:hover { background: #e0a030; }
		.back-link {
			display: block;
			text-align: center;
			margin-top: 15px;
		}
	</style>
</head>
<body>
	<h1>Agregar Nueva Serie</h1>
	<form method="POST" action="/create">
		<label>Nombre de la serie:</label>
		<input type="text" name="series_name" required>

		<label>Episodio actual:</label>
		<input type="number" name="current_episode" min="0" value="1" required>

		<label>Total de episodios:</label>
		<input type="number" name="total_episodes" min="1" required>

		<button type="submit">Agregar Serie</button>
	</form>
	<a class="back-link" href="/">Volver a la lista</a>
</body>
</html>
`, sharedCSS)
}
