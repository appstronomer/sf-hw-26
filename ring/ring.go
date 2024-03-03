package ring

func MakeRingBufferLoop[T any](size int) func(chIn <-chan T, chOut chan<- T) {
	return func(chIn <-chan T, chOut chan<- T) {
		sl := make([]T, size)
		idxSet := 0
		idxRm := 0
		var val T
		var ok bool

		// получение первого сообщения
		val, ok = <-chIn
		if !ok {
			// ни одного сообщения не получено - можно выходить
			return
		}
		sl[idxSet] = val
		idxSet = (idxSet + 1) % size

		for {
			select {

			case val, ok = <-chIn:
				// получено сообщение через входной канал
				if !ok {
					// Канал закрыт
					if idxSet == idxRm {
						// idxSet сделал круг и догнал idxRm
						chOut <- sl[idxRm]
						idxRm = (idxRm + 1) % size
					}
					for idxRm != idxSet {
						// отправка неотправленных сообщений
						chOut <- sl[idxRm]
						idxRm = (idxRm + 1) % size
					}
					return
				}
				if idxSet == idxRm {
					idxRm = (idxRm + 1) % size
				}
				sl[idxSet] = val
				idxSet = (idxSet + 1) % size

			case chOut <- sl[idxRm]:
				// отправлено сообщение через выходной канал
				idxRm = (idxRm + 1) % size
				if idxRm == idxSet {
					// новых сообщений нет
					val, ok = <-chIn
					if !ok {
						// все сообщения были отправлены - можно выходить
						return
					}
					sl[idxSet] = val
					idxSet = (idxSet + 1) % size
				}
			}
		}

	}
}
