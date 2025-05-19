package batcher

func Split[T any](items []T, size int) [][]T {
    var batches [][]T
    for size < len(items) {
        items, batches = items[size:], append(batches, items[0:size:size])
    }
    return append(batches, items)
}