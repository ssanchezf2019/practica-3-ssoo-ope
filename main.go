package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	torre := TorreControl{
		pistas:       make(chan struct{}, NumPistas),
		puertas:      make(chan struct{}, NumPuertas),
		inicio:       time.Now(),
		wg:           &sync.WaitGroup{},
		messageMutex: sync.Mutex{},
		ordenMutex:   sync.Mutex{},
	}
	torre.cond = sync.NewCond(&torre.mutex)

	// Crear aviones
	for i := 0; i < NumAviones; i++ {
		torre.wg.Add(1)
		go func(id int) {
			numPasajeros := rand.Intn(150) + 1
			categoria := calcularCategoria(numPasajeros)
			avion := Avion{
				id:           id,
				numPasajeros: numPasajeros,
				categoria:    categoria,
				horaConexion: time.Now(),
			}
			torre.encolarAvion(avion)
		}(i)
	}

	// Procesar aviones
	go torre.gestionarColas()

	// Esperar a que todos los aviones terminen
	torre.wg.Wait()

	// Calcular y mostrar resultados
	torre.log("\nSimulación completada\n")
	torre.log(fmt.Sprintf("Promedio de tiempo desde conexión hasta aterrizaje: %.2f segundos\n", calcularPromedio(torre.tiemposConexiónAterrizaje)))
	torre.log(fmt.Sprintf("Promedio de tiempo desde aterrizaje hasta desembarque: %.2f segundos\n", calcularPromedio(torre.tiemposAterrizajeDesembarque)))
}

// Calcular categoría según número de pasajeros
func calcularCategoria(numPasajeros int) string {
	if numPasajeros > 100 {
		return "A"
	} else if numPasajeros >= 50 {
		return "B"
	}
	return "C"
}

// Método para encolar un avión
func (t *TorreControl) encolarAvion(avion Avion) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if len(t.esperaActiva) < MaxEspera {
		t.insertarYReordenar(&t.esperaActiva, avion)
		t.log(fmt.Sprintf(msgConectado, t.tiempoTranscurrido(), avion.id, avion.categoria, avion.numPasajeros))
		t.cond.Signal() // Notificar que hay un avión nuevo en espera activa
	} else {
		t.insertarYReordenar(&t.fueraDeEspera, avion)
		t.log(fmt.Sprintf(msgEsperandoFuera, t.tiempoTranscurrido(), avion.id, avion.categoria, avion.numPasajeros))
	}
}

// Método para gestionar las colas de aviones
func (t *TorreControl) gestionarColas() {
	for {
		// Reservar una pista (bloquea si no hay disponibles)
		t.pistas <- struct{}{}

		t.mutex.Lock()

		// Mover aviones de fueraDeEspera a esperaActiva si hay espacio
		for len(t.esperaActiva) < MaxEspera && len(t.fueraDeEspera) > 0 {
			avion := t.extraerPrioritario(&t.fueraDeEspera)
			t.insertarYReordenar(&t.esperaActiva, avion)
			t.log(fmt.Sprintf(msgConectado, t.tiempoTranscurrido(), avion.id, avion.categoria, avion.numPasajeros))
		}

		// Seleccionar un avión para aterrizar, si hay aviones en espera
		if len(t.esperaActiva) > 0 {
			avion := t.extraerPrioritario(&t.esperaActiva)
			t.log(fmt.Sprintf(msgAterrizando, t.tiempoTranscurrido(), avion.id, avion.categoria))

			t.mutex.Unlock()

			// Procesar el aterrizaje
			go func(avion Avion) {
				t.procesarAterrizaje(avion)
				<-t.pistas // Liberar la pista después de aterrizar
			}(avion)
		} else {
			// No hay aviones en espera; liberar la pista
			<-t.pistas
			t.cond.Wait() // Esperar notificación de nuevos aviones
			t.mutex.Unlock()
		}
	}
}

// Método para procesar aterrizaje
func (t *TorreControl) procesarAterrizaje(avion Avion) {
	defer t.wg.Done()

	// Simular el aterrizaje
	time.Sleep(tiempoConVariacion(TiempoAterrizaje))

	tiempoAterrizaje := time.Since(avion.horaConexion).Seconds()
	t.log(fmt.Sprintf(msgAterrizado, t.tiempoTranscurrido(), avion.id, avion.categoria, tiempoAterrizaje))

	t.mutex.Lock()
	t.tiemposConexiónAterrizaje = append(t.tiemposConexiónAterrizaje, tiempoAterrizaje)
	t.mutex.Unlock()

	// Procesar desembarque
	t.procesarDesembarque(avion)
}

// Método para procesar desembarque
func (t *TorreControl) procesarDesembarque(avion Avion) {
	t.puertas <- struct{}{}
	t.log(fmt.Sprintf(msgEnPuerta, t.tiempoTranscurrido(), avion.id, avion.categoria))

	startTime := time.Now()
	time.Sleep(tiempoDesembarquePorCategoria(avion.categoria))

	tiempoDesembarque := time.Since(startTime).Seconds()
	t.log(fmt.Sprintf(msgTerminoDesembarque, t.tiempoTranscurrido(), avion.id, avion.categoria, tiempoDesembarque))

	t.mutex.Lock()
	t.tiemposAterrizajeDesembarque = append(t.tiemposAterrizajeDesembarque, tiempoDesembarque)
	t.mutex.Unlock()

	<-t.puertas
	t.cond.Signal()
}

// Insertar un avión y reordenar la lista según prioridad
func (t *TorreControl) insertarYReordenar(lista *[]Avion, avion Avion) {
	*lista = append(*lista, avion)
	sort.SliceStable(*lista, func(i, j int) bool {
		return (*lista)[i].categoria < (*lista)[j].categoria
	})
}

// Extraer avión con mayor prioridad
func (t *TorreControl) extraerPrioritario(lista *[]Avion) Avion {
	avion := (*lista)[0]
	*lista = (*lista)[1:]
	return avion
}

// Método para imprimir mensajes sincronizados
func (t *TorreControl) log(message string) {
	t.messageMutex.Lock()
	defer t.messageMutex.Unlock()
	fmt.Print(message)
}

// Calcular tiempo promedio
func calcularPromedio(tiempos []float64) float64 {
	if len(tiempos) == 0 {
		return 0
	}
	var suma float64
	for _, tiempo := range tiempos {
		suma += tiempo
	}
	return suma / float64(len(tiempos))
}

// Obtener tiempo transcurrido
func (t *TorreControl) tiempoTranscurrido() string {
	return fmt.Sprintf("%.2fs", time.Since(t.inicio).Seconds())
}

// Variación en tiempos
func tiempoConVariacion(base int) time.Duration {
	variacion := base * VariacionTiempo / 100
	return time.Duration(base+rand.Intn(2*variacion)-variacion) * time.Millisecond
}

// Tiempo de puerta personalizado por categoría
func tiempoDesembarquePorCategoria(categoria string) time.Duration {
	switch categoria {
	case "A":
		return tiempoConVariacion(TiempoPuerta + 2000)
	case "B":
		return tiempoConVariacion(TiempoPuerta + 1000)
	default:
		return tiempoConVariacion(TiempoPuerta)
	}
}
