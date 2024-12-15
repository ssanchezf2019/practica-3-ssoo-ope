// EJECUCIÓN: go test -v

package main

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

var resultadosTests []resultadoTest

// Función de ayuda para ejecutar la simulación y calcular los promedios
func ejecutarSimulacionPorCategorias(avionesA, avionesB, avionesC int) (float64, float64) {
	torre := TorreControl{
		pistas:       make(chan struct{}, NumPistas),
		puertas:      make(chan struct{}, NumPuertas),
		inicio:       time.Now(),
		wg:           &sync.WaitGroup{},
		messageMutex: sync.Mutex{},
		ordenMutex:   sync.Mutex{},
	}
	torre.cond = sync.NewCond(&torre.mutex)

	// Crear un slice para los aviones y asignar categorías mezcladas
	totalAviones := avionesA + avionesB + avionesC
	tiposAviones := make([]string, 0, totalAviones)

	// Agregar los tipos de avión al slice
	for i := 0; i < avionesA; i++ {
		tiposAviones = append(tiposAviones, "A")
	}
	for i := 0; i < avionesB; i++ {
		tiposAviones = append(tiposAviones, "B")
	}
	for i := 0; i < avionesC; i++ {
		tiposAviones = append(tiposAviones, "C")
	}

	// Mezclar aleatoriamente las categorías
	rand.Shuffle(len(tiposAviones), func(i, j int) { tiposAviones[i], tiposAviones[j] = tiposAviones[j], tiposAviones[i] })

	// Crear los aviones mezclados
	for i, categoria := range tiposAviones {
		torre.wg.Add(1)
		numPasajeros := 0
		switch categoria {
		case "A":
			numPasajeros = 120 // Pasajeros típicos para categoría A
		case "B":
			numPasajeros = 80 // Pasajeros típicos para categoría B
		case "C":
			numPasajeros = 40 // Pasajeros típicos para categoría C
		}
		avion := Avion{id: i, categoria: categoria, numPasajeros: numPasajeros, horaConexion: time.Now()}
		go torre.encolarAvion(avion)
	}

	// Procesar la cola y gestionar la espera activa
	go torre.gestionarColas()

	// Esperar a que todos los aviones terminen
	torre.wg.Wait()

	// Calcular los promedios
	promedioConexionAterrizaje := calcularPromedio(torre.tiemposConexiónAterrizaje)
	promedioAterrizajeDesembarque := calcularPromedio(torre.tiemposAterrizajeDesembarque)

	return promedioConexionAterrizaje, promedioAterrizajeDesembarque
}

// TestMain para ejecutar todos los tests y luego mostrar la tabla comparativa
func TestMain(m *testing.M) {
	// Ejecuta todas las pruebas
	m.Run()

	// Generar tabla comparativa de los resultados
	fmt.Printf("\nTabla Comparativa de Tiempos Promedio\n")
	fmt.Printf("%-50s | %-40s | %-40s\n", "Prueba", "Promedio Conexión a Aterrizaje (s)", "Promedio Aterrizaje a Desembarque (s)")
	fmt.Println(strings.Repeat("-", 110))
	for _, res := range resultadosTests {
		fmt.Printf("%-50s | %-40.2f | %-40.2f\n", res.nombre, res.promedioConexionAterrizaje, res.promedioAterrizajeDesembarque)
	}
}

// Función para almacenar resultados de cada prueba en la tabla comparativa
func registrarResultados(nombre string, promedioConexionAterrizaje, promedioAterrizajeDesembarque float64) {
	resultadosTests = append(resultadosTests, resultadoTest{
		nombre:                        nombre,
		promedioConexionAterrizaje:    promedioConexionAterrizaje,
		promedioAterrizajeDesembarque: promedioAterrizajeDesembarque,
	})
}

// Prueba 1: 10 aviones categoría A, 10 categoría B, 10 categoría C
func TestCaso1(t *testing.T) {
	promedioConexionAterrizaje, promedioAterrizajeDesembarque := ejecutarSimulacionPorCategorias(10, 10, 10)

	registrarResultados("Caso 1: 10 A, 10 B, 10 C", promedioConexionAterrizaje, promedioAterrizajeDesembarque)
}

// Prueba 2: 20 aviones categoría A, 5 categoría B, 5 categoría C
func TestCaso2(t *testing.T) {
	promedioConexionAterrizaje, promedioAterrizajeDesembarque := ejecutarSimulacionPorCategorias(20, 5, 5)

	registrarResultados("Caso 2: 20 A, 5 B, 5 C", promedioConexionAterrizaje, promedioAterrizajeDesembarque)
}

// Prueba 3: 5 aviones categoría A, 5 categoría B, 20 categoría C
func TestCaso3(t *testing.T) {
	promedioConexionAterrizaje, promedioAterrizajeDesembarque := ejecutarSimulacionPorCategorias(5, 5, 20)

	registrarResultados("Caso 3: 5 A, 5 B, 20 C", promedioConexionAterrizaje, promedioAterrizajeDesembarque)
}
