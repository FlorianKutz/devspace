package server

import (
	"github.com/loft-sh/devspace/pkg/devspace/kubectl"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/remotecommand"
	"net/http"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gorilla/websocket"
)

func (h *handler) enter(w http.ResponseWriter, r *http.Request) {
	// Kube Context
	kubeContext := h.defaultContext
	context, ok := r.URL.Query()["context"]
	if ok && len(context) == 1 && context[0] != "" {
		kubeContext = context[0]
	}

	// Namespace
	kubeNamespace := h.defaultNamespace
	namespace, ok := r.URL.Query()["namespace"]
	if ok && len(namespace) == 1 && namespace[0] != "" {
		kubeNamespace = namespace[0]
	}

	name, ok := r.URL.Query()["name"]
	if !ok || len(name) != 1 {
		http.Error(w, "name is missing", http.StatusBadRequest)
		return
	}
	container, ok := r.URL.Query()["container"]
	if !ok || len(container) != 1 {
		http.Error(w, "container is missing", http.StatusBadRequest)
		return
	}

	var terminalResizeQueue TerminalResizeQueue
	resizeId, ok := r.URL.Query()["resize_id"]
	if ok && len(resizeId) == 1 {
		h.terminalResizeQueuesMutex.Lock()
		terminalResizeQueue = newTerminalSizeQueue()
		h.terminalResizeQueues[resizeId[0]] = terminalResizeQueue
		h.terminalResizeQueuesMutex.Unlock()

		defer func() {
			h.terminalResizeQueuesMutex.Lock()
			defer h.terminalResizeQueuesMutex.Unlock()

			delete(h.terminalResizeQueues, resizeId[0])
		}()
	}

	// Create kubectl client
	client, err := h.getClientFromCache(kubeContext, kubeNamespace)
	if err != nil {
		h.log.Errorf("Error in %s: %v", r.URL.String(), err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Errorf("Error upgrading connection: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer ws.Close()

	// Open logs connection
	stream := &wsStream{WebSocket: ws}
	err = client.ExecStream(&kubectl.ExecStreamOptions{
		Pod: &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name[0],
				Namespace: namespace[0],
			},
		},
		Container:         container[0],
		Command:           []string{"sh", "-c", "command -v bash >/dev/null 2>&1 && exec bash || exec sh"},
		ForceTTY:          true,
		TTY:               true,
		TerminalSizeQueue: terminalResizeQueue,
		Stdin:             stream,
		Stdout:            stream,
		Stderr:            stream,
	})
	if err != nil {
		h.log.Errorf("Error in %s: %v", r.URL.String(), err)
		websocketError(ws, err)
		return
	}

	ws.SetWriteDeadline(time.Now().Add(time.Second * 5))
	ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (h *handler) resize(w http.ResponseWriter, r *http.Request) {
	resizeId, ok := r.URL.Query()["resize_id"]
	if !ok || len(resizeId) != 1 {
		http.Error(w, "resize_id is missing", http.StatusBadRequest)
		return
	}
	widthStr, ok := r.URL.Query()["width"]
	if !ok || len(widthStr) != 1 {
		http.Error(w, "width is missing", http.StatusBadRequest)
		return
	}
	heightStr, ok := r.URL.Query()["height"]
	if !ok || len(heightStr) != 1 {
		http.Error(w, "height is missing", http.StatusBadRequest)
		return
	}
	width, err := strconv.Atoi(widthStr[0])
	if err != nil {
		http.Error(w, errors.Wrap(err, "parse width").Error(), http.StatusBadRequest)
		return
	}
	height, err := strconv.Atoi(heightStr[0])
	if err != nil {
		http.Error(w, errors.Wrap(err, "parse height").Error(), http.StatusBadRequest)
		return
	}

	h.terminalResizeQueuesMutex.Lock()
	defer h.terminalResizeQueuesMutex.Unlock()

	resizeQueue, ok := h.terminalResizeQueues[resizeId[0]]
	if !ok {
		http.Error(w, "resize_id does not exist", http.StatusBadRequest)
		return
	}

	resizeQueue.Resize(remotecommand.TerminalSize{
		Width:  uint16(width),
		Height: uint16(height),
	})
}

type TerminalResizeQueue interface {
	remotecommand.TerminalSizeQueue

	Resize(size remotecommand.TerminalSize)
}

func newTerminalSizeQueue() TerminalResizeQueue {
	return &terminalSizeQueue{
		resizeChan: make(chan remotecommand.TerminalSize, 1),
	}
}

type terminalSizeQueue struct {
	resizeChan chan remotecommand.TerminalSize
}

func (t *terminalSizeQueue) Resize(size remotecommand.TerminalSize) {
	select {
	// try to send the size to resizeChan, but don't block
	case t.resizeChan <- size:
		// send successful
	default:
		// unable to send / no-op
	}
}

func (t *terminalSizeQueue) Next() *remotecommand.TerminalSize {
	size, ok := <-t.resizeChan
	if !ok {
		return nil
	}
	return &size
}
