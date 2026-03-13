# Sistemas y Tecnologías Web - Ejercicio 4

API para servir inforamción sobre países. 

## Levantar el proyecto

```bash
docker compose up --build
```

La API queda disponible en `http://localhost:24880`.

## Endpoints

### `GET /api/ping`
Verifica que el servidor está en línea.

### `GET /api/countries`
Devuelve todos los países.

### `GET /api/countries?id=1`
Devuelve un país por ID.

### `Delete /api/countries?id=1`
Elimina un país por ID.

### `POST /api/countries`
Crea un nuevo país. Todos los campos son requeridos.

**Errores:**
* 400:  `id` no es un número válido 
* 404:  País no encontrado 



## Modelo

```json
{
  "id": 1,
  "name": "string",
  "capital": "string",
  "population": 0,
  "continent": "string",
  "currency": "string"
}
```