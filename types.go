package main

import (
	"sync"
	"time"
)

// Representa un avión en el sistema
type Avion struct {
	id           int       // Identificador único del avión
	numPasajeros int       // Número de pasajeros
	categoria    string    // Categoría (A, B o C) para priorización
	horaConexion time.Time // Hora de conexión a la torre de control
}

// TorreControl representa el control de tráfico aéreo
type TorreControl struct {
	pistas                       chan struct{}
	puertas                      chan struct{}
	esperaActiva                 []Avion
	fueraDeEspera                []Avion
	tiemposConexiónAterrizaje    []float64
	tiemposAterrizajeDesembarque []float64
	mutex                        sync.Mutex
	cond                         *sync.Cond
	wg                           *sync.WaitGroup
	inicio                       time.Time
	messageMutex                 sync.Mutex // Mutex exclusivo para mensajes
	ordenMutex                   sync.Mutex // Mutex exclusivo para orden de aterrizaje
}

// Estructura para almacenar los resultados de cada prueba
type resultadoTest struct {
	nombre                        string
	promedioConexionAterrizaje    float64
	promedioAterrizajeDesembarque float64
}
