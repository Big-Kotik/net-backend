package com.kotik.big.netbackend

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication

@SpringBootApplication
class NetBackendApplication

fun main(args: Array<String>) {
	runApplication<NetBackendApplication>(*args)
}
