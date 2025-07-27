// static/js/export.js - Sistema Export Completo
class ExportManager {
    constructor() {
        this.exportRooms = [];
        this.currentPreview = null;
        this.init();
    }

    async init() {
        await this.loadExportRooms();
        this.setupEventListeners();
    }

    async loadExportRooms() {
        try {
            const response = await fetch('/api/rooms', {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });
            this.exportRooms = await response.json();

            const select = document.getElementById('exportRoom');
            if (select) {
                select.innerHTML = '<option value="">Tutte le stanze</option>';
                this.exportRooms.forEach(room => {
                    const option = document.createElement('option');
                    option.value = room.id;
                    option.textContent = room.name;
                    select.appendChild(option);
                });
            }
        } catch (error) {
            console.error('Errore nel caricamento stanze per export:', error);
        }
    }

    setupEventListeners() {
        // Event listener per cambio tipo export
        const exportType = document.getElementById('exportType');
        if (exportType) {
            exportType.addEventListener('change', () => this.updateExportFilters());
        }

        // Event listener per form export
        const exportForm = document.getElementById('exportForm');
        if (exportForm) {
            exportForm.addEventListener('submit', (e) => this.handleExportSubmit(e));
        }

        // Event listeners per preview in tempo reale
        const filterInputs = ['exportStartDate', 'exportEndDate', 'exportStatus', 'exportRoom'];
        filterInputs.forEach(id => {
            const element = document.getElementById(id);
            if (element) {
                element.addEventListener('change', () => this.updatePreview());
            }
        });
    }

    openExportModal(defaultType = '') {
        const modal = document.getElementById('exportModal');
        if (!modal) return;

        modal.classList.remove('hidden');
        modal.classList.add('flex');

        if (defaultType) {
            document.getElementById('exportType').value = defaultType;
            this.updateExportFilters();
        }

        // Imposta date di default (ultimo mese)
        const endDate = new Date();
        const startDate = new Date();
        startDate.setMonth(startDate.getMonth() - 1);

        document.getElementById('exportStartDate').value = startDate.toISOString().split('T')[0];
        document.getElementById('exportEndDate').value = endDate.toISOString().split('T')[0];

        // Reset form
        this.resetExportForm();
    }

    closeExportModal() {
        const modal = document.getElementById('exportModal');
        if (modal) {
            modal.classList.add('hidden');
            modal.classList.remove('flex');
        }

        this.resetExportForm();
        this.hideLoading();
    }

    resetExportForm() {
        const form = document.getElementById('exportForm');
        if (form) {
            form.reset();
        }

        const preview = document.getElementById('exportPreview');
        if (preview) {
            preview.classList.add('hidden');
        }
    }

    updateExportFilters() {
        const type = document.getElementById('exportType').value;
        const statusFilter = document.getElementById('statusFilter');
        const roomFilter = document.getElementById('roomFilter');
        const exportStatus = document.getElementById('exportStatus');

        // Nascondi tutti i filtri per default
        statusFilter.classList.add('hidden');
        roomFilter.classList.add('hidden');

        // Mostra filtri appropriati
        switch (type) {
            case 'hedgehogs':
                statusFilter.classList.remove('hidden');
                roomFilter.classList.remove('hidden');
                exportStatus.innerHTML = `
                    <option value="">Tutti gli stati</option>
                    <option value="in_care">In cura</option>
                    <option value="recovered">Recuperato</option>
                    <option value="deceased">Deceduto</option>
                `;
                break;
            case 'therapies':
                statusFilter.classList.remove('hidden');
                exportStatus.innerHTML = `
                    <option value="">Tutti gli stati</option>
                    <option value="active">Attiva</option>
                    <option value="completed">Completata</option>
                    <option value="suspended">Sospesa</option>
                `;
                break;
            case 'rooms':
                roomFilter.classList.remove('hidden');
                break;
        }

        // Aggiorna preview
        this.updatePreview();
    }

    async updatePreview() {
        const type = document.getElementById('exportType').value;
        if (!type) return;

        const preview = document.getElementById('exportPreview');
        const previewContent = document.getElementById('previewContent');

        if (!preview || !previewContent) return;

        try {
            const filters = this.getExportFilters();
            const stats = await this.getExportStats(type, filters);

            previewContent.innerHTML = `
                <div class="preview-item">
                    <span>Tipo dati:</span>
                    <span>${this.getTypeLabel(type)}</span>
                </div>
                <div class="preview-item">
                    <span>Periodo:</span>
                    <span>${this.formatDateRange(filters.start_date, filters.end_date)}</span>
                </div>
                ${filters.status ? `
                <div class="preview-item">
                    <span>Stato:</span>
                    <span>${filters.status}</span>
                </div>` : ''}
                ${filters.room_id ? `
                <div class="preview-item">
                    <span>Stanza:</span>
                    <span>${this.getRoomName(filters.room_id)}</span>
                </div>` : ''}
                <div class="preview-item">
                    <span>Record da esportare:</span>
                    <span>${stats.count}</span>
                </div>
            `;

            preview.classList.remove('hidden');
        } catch (error) {
            console.error('Errore nel preview:', error);
        }
    }

    async getExportStats(type, filters) {
        // Simulazione conteggio per preview
        // In produzione fare una chiamata API dedicata
        return { count: Math.floor(Math.random() * 100) + 1 };
    }

    getExportFilters() {
        return {
            start_date: document.getElementById('exportStartDate').value,
            end_date: document.getElementById('exportEndDate').value,
            status: document.getElementById('exportStatus').value,
            room_id: document.getElementById('exportRoom').value
        };
    }

    getTypeLabel(type) {
        const labels = {
            'hedgehogs': '🦔 Ricci',
            'rooms': '🏠 Stanze e Aree',
            'therapies': '💊 Terapie',
            'weights': '⚖️ Pesature'
        };
        return labels[type] || type;
    }

    formatDateRange(startDate, endDate) {
        if (!startDate && !endDate) return 'Tutto il periodo';
        if (!startDate) return `Fino al ${new Date(endDate).toLocaleDateString('it-IT')}`;
        if (!endDate) return `Dal ${new Date(startDate).toLocaleDateString('it-IT')}`;
        return `${new Date(startDate).toLocaleDateString('it-IT')} - ${new Date(endDate).toLocaleDateString('it-IT')}`;
    }

    getRoomName(roomId) {
        const room = this.exportRooms.find(r => r.id == roomId);
        return room ? room.name : 'Sconosciuta';
    }

    async handleExportSubmit(e) {
        e.preventDefault();

        const formData = new FormData(e.target);
        const exportData = {
            type: formData.get('type'),
            format: formData.get('format')
        };

        // Aggiungi filtri opzionali
        const filters = this.getExportFilters();
        Object.keys(filters).forEach(key => {
            if (filters[key]) {
                if (key === 'room_id') {
                    exportData[key] = parseInt(filters[key]);
                } else {
                    exportData[key] = filters[key];
                }
            }
        });

        await this.performExport(exportData);
    }

    async performExport(exportData) {
        this.showLoading();

        try {
            const response = await fetch('/api/export', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(exportData)
            });

            if (response.ok) {
                await this.downloadFile(response, exportData);
                this.closeExportModal();
                this.showNotification('Export completato con successo!', 'success');
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Errore durante l\'export');
            }
        } catch (error) {
            console.error('Errore export:', error);
            this.showNotification('Errore durante l\'export: ' + error.message, 'error');
            this.hideLoading();
        }
    }

    async downloadFile(response, exportData) {
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;

        // Estrai il nome del file dall'header Content-Disposition
        const contentDisposition = response.headers.get('Content-Disposition');
        let filename = `la-ninna-${exportData.type}-${new Date().toISOString().split('T')[0]}.${exportData.format}`;
        if (contentDisposition) {
            const matches = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(contentDisposition);
            if (matches != null && matches[1]) {
                filename = matches[1].replace(/['"]/g, '');
            }
        }

        a.download = filename;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
    }

    showLoading() {
        const form = document.getElementById('exportForm');
        const loading = document.getElementById('exportLoading');
        const progress = document.getElementById('exportProgress');

        if (form) form.classList.add('hidden');
        if (loading) loading.classList.remove('hidden');

        // Simula progress bar
        let width = 0;
        const interval = setInterval(() => {
            width += Math.random() * 30;
            if (width >= 100) {
                width = 100;
                clearInterval(interval);
            }
            if (progress) {
                progress.style.width = width + '%';
            }
        }, 500);

        this.progressInterval = interval;
    }

    hideLoading() {
        const form = document.getElementById('exportForm');
        const loading = document.getElementById('exportLoading');

        if (loading) loading.classList.add('hidden');
        if (form) form.classList.remove('hidden');

        if (this.progressInterval) {
            clearInterval(this.progressInterval);
        }
    }

    async quickExport(type, format, filters = {}) {
        const exportData = { type, format, ...filters };

        try {
            this.showNotification('Generazione in corso...', 'info');

            const response = await fetch('/api/export', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(exportData)
            });

            if (response.ok) {
                await this.downloadFile(response, exportData);
                this.showNotification('Export completato!', 'success');
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Errore durante l\'export');
            }
        } catch (error) {
            console.error('Errore quick export:', error);
            this.showNotification('Errore durante l\'export: ' + error.message, 'error');
        }
    }

    showNotification(message, type = 'info') {
        // Rimuovi notifiche esistenti
        const existingNotifications = document.querySelectorAll('.notification');
        existingNotifications.forEach(n => n.remove());

        const notification = document.createElement('div');
        notification.className = `notification fixed top-4 right-4 z-50 px-6 py-4 rounded-lg shadow-lg transform transition-all duration-300 translate-x-full`;

        const colors = {
            success: 'bg-green-500 text-white',
            error: 'bg-red-500 text-white',
            info: 'bg-blue-500 text-white',
            warning: 'bg-yellow-500 text-black'
        };

        const icons = {
            success: 'fas fa-check-circle',
            error: 'fas fa-exclamation-circle',
            info: 'fas fa-info-circle',
            warning: 'fas fa-exclamation-triangle'
        };

        notification.className += ` ${colors[type]}`;
        notification.innerHTML = `
            <div class="flex items-center">
                <i class="${icons[type]} mr-3"></i>
                <span>${message}</span>
                <button onclick="this.parentElement.parentElement.remove()" class="ml-4 hover:opacity-70">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        `;

        document.body.appendChild(notification);

        // Animazione di entrata
        setTimeout(() => {
            notification.classList.remove('translate-x-full');
        }, 100);

        // Auto-remove dopo 5 secondi
        setTimeout(() => {
            notification.classList.add('translate-x-full');
            setTimeout(() => {
                if (notification.parentElement) {
                    notification.remove();
                }
            }, 300);
        }, 5000);
    }

    // Metodi pubblici per compatibilità globale
    async previewExport() {
        this.updatePreview();

        // Scroll verso il preview
        const preview = document.getElementById('exportPreview');
        if (preview && !preview.classList.contains('hidden')) {
            preview.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        }
    }
}

// Inizializza export manager globalmente
let exportManager;

document.addEventListener('DOMContentLoaded', function() {
    exportManager = new ExportManager();
});

// Funzioni globali per compatibilità con template esistenti
function openExportModal(defaultType = '') {
    if (exportManager) {
        exportManager.openExportModal(defaultType);
    }
}

function closeExportModal() {
    if (exportManager) {
        exportManager.closeExportModal();
    }
}

function quickExport(type, format, filters = {}) {
    if (exportManager) {
        return exportManager.quickExport(type, format, filters);
    }
}

function previewExport() {
    if (exportManager) {
        exportManager.previewExport();
    }
}

function updateExportFilters() {
    if (exportManager) {
        exportManager.updateExportFilters();
    }
}

function showNotification(message, type = 'info') {
    if (exportManager) {
        exportManager.showNotification(message, type);
    }
}

// Export per URL diretti (per link rapidi)
function exportHedgehogsPDF() {
    window.open('/api/export/hedgehogs/pdf', '_blank');
}

function exportHedgehogsExcel() {
    window.open('/api/export/hedgehogs/excel', '_blank');
}

function exportRoomsPDF() {
    window.open('/api/export/rooms/pdf', '_blank');
}

function exportRoomsExcel() {
    window.open('/api/export/rooms/excel', '_blank');
}

function exportTherapiesPDF() {
    window.open('/api/export/therapies/pdf', '_blank');
}

function exportTherapiesExcel() {
    window.open('/api/export/therapies/excel', '_blank');
}

function exportWeightsPDF() {
    window.open('/api/export/weights/pdf', '_blank');
}

function exportWeightsExcel() {
    window.open('/api/export/weights/excel', '_blank');
}

// Utilità per export batch (esporta tutti i formati)
async function exportAll(type) {
    if (!exportManager) return;

    const formats = ['pdf', 'excel', 'csv'];
    exportManager.showNotification('Avvio export multiplo...', 'info');

    for (const format of formats) {
        try {
            await exportManager.quickExport(type, format);
            await new Promise(resolve => setTimeout(resolve, 1000)); // Pausa tra export
        } catch (error) {
            console.error(`Errore export ${format}:`, error);
        }
    }

    exportManager.showNotification('Export multiplo completato!', 'success');
}

// Export programmato (per future implementazioni)
class ScheduledExport {
    constructor() {
        this.schedules = JSON.parse(localStorage.getItem('export_schedules') || '[]');
    }

    addSchedule(config) {
        const schedule = {
            id: Date.now(),
            type: config.type,
            format: config.format,
            filters: config.filters || {},
            frequency: config.frequency, // daily, weekly, monthly
            nextRun: this.calculateNextRun(config.frequency),
            enabled: true,
            created: new Date().toISOString()
        };

        this.schedules.push(schedule);
        this.saveSchedules();
        return schedule;
    }

    removeSchedule(id) {
        this.schedules = this.schedules.filter(s => s.id !== id);
        this.saveSchedules();
    }

    calculateNextRun(frequency) {
        const now = new Date();
        switch (frequency) {
            case 'daily':
                now.setDate(now.getDate() + 1);
                break;
            case 'weekly':
                now.setDate(now.getDate() + 7);
                break;
            case 'monthly':
                now.setMonth(now.getMonth() + 1);
                break;
        }
        return now.toISOString();
    }

    saveSchedules() {
        localStorage.setItem('export_schedules', JSON.stringify(this.schedules));
    }

    async runDueSchedules() {
        const now = new Date();
        const dueSchedules = this.schedules.filter(s =>
            s.enabled && new Date(s.nextRun) <= now
        );

        for (const schedule of dueSchedules) {
            try {
                await exportManager.quickExport(schedule.type, schedule.format, schedule.filters);

                // Aggiorna prossima esecuzione
                schedule.nextRun = this.calculateNextRun(schedule.frequency);
                schedule.lastRun = now.toISOString();

                console.log(`Export programmato eseguito: ${schedule.type} ${schedule.format}`);
            } catch (error) {
                console.error('Errore export programmato:', error);
                schedule.lastError = error.message;
            }
        }

        if (dueSchedules.length > 0) {
            this.saveSchedules();
        }
    }

    getSchedules() {
        return this.schedules;
    }
}

// Inizializza export programmati
const scheduledExport = new ScheduledExport();

// Controlla export programmati ogni ora
setInterval(() => {
    scheduledExport.runDueSchedules();
}, 60 * 60 * 1000);

// Export con compressione per grandi dataset
class CompressedExport {
    static async exportLarge(type, format, filters = {}) {
        try {
            exportManager.showNotification('Preparazione export di grandi dimensioni...', 'info');

            // Per dataset molto grandi, implementa chunking
            const chunkSize = 1000;
            let offset = 0;
            const chunks = [];

            while (true) {
                const chunkFilters = {
                    ...filters,
                    limit: chunkSize,
                    offset: offset
                };

                const response = await fetch('/api/export', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        type,
                        format: 'json', // Formato intermedio
                        ...chunkFilters
                    })
                });

                if (!response.ok) break;

                const chunk = await response.json();
                if (chunk.length === 0) break;

                chunks.push(chunk);
                offset += chunkSize;

                // Aggiorna progress
                exportManager.showNotification(`Processati ${offset} record...`, 'info');
            }

            // Combina chunks e esporta
            const combinedData = chunks.flat();
            return await this.exportCombined(combinedData, type, format);

        } catch (error) {
            console.error('Errore export grande:', error);
            exportManager.showNotification('Errore export grandi dimensioni: ' + error.message, 'error');
        }
    }

    static async exportCombined(data, type, format) {
        // Implementa export combinato locale per grandi dataset
        // Questo riduce il carico sul server

        if (format === 'csv') {
            return this.exportToCSV(data, type);
        } else if (format === 'json') {
            return this.exportToJSON(data, type);
        }

        // Per PDF/Excel, invia al server in chunks più piccoli
        return exportManager.quickExport(type, format);
    }

    static exportToCSV(data, type) {
        // Implementazione export CSV lato client
        const csv = this.jsonToCSV(data);
        const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);

        const a = document.createElement('a');
        a.href = url;
        a.download = `la-ninna-${type}-large-${new Date().toISOString().split('T')[0]}.csv`;
        a.click();

        URL.revokeObjectURL(url);
        exportManager.showNotification('Export CSV completato!', 'success');
    }

    static exportToJSON(data, type) {
        const json = JSON.stringify(data, null, 2);
        const blob = new Blob([json], { type: 'application/json;charset=utf-8;' });
        const url = URL.createObjectURL(blob);

        const a = document.createElement('a');
        a.href = url;
        a.download = `la-ninna-${type}-large-${new Date().toISOString().split('T')[0]}.json`;
        a.click();

        URL.revokeObjectURL(url);
        exportManager.showNotification('Export JSON completato!', 'success');
    }

    static jsonToCSV(data) {
        if (!data || data.length === 0) return '';

        const headers = Object.keys(data[0]);
        const csvHeaders = headers.join(';');

        const csvRows = data.map(row => {
            return headers.map(header => {
                const value = row[header];
                // Escape per CSV
                if (typeof value === 'string' && (value.includes(';') || value.includes('"') || value.includes('\n'))) {
                    return '"' + value.replace(/"/g, '""') + '"';
                }
                return value || '';
            }).join(';');
        });

        return [csvHeaders, ...csvRows].join('\n');
    }
}

// Esporta classi per uso globale
window.CompressedExport = CompressedExport;
window.ScheduledExport = ScheduledExport;

// Analytics per export (tracking utilizzo)
class ExportAnalytics {
    static track(type, format, success = true) {
        const stats = JSON.parse(localStorage.getItem('export_stats') || '{}');
        const key = `${type}_${format}`;

        if (!stats[key]) {
            stats[key] = { count: 0, success: 0, lastUsed: null };
        }

        stats[key].count++;
        if (success) stats[key].success++;
        stats[key].lastUsed = new Date().toISOString();

        localStorage.setItem('export_stats', JSON.stringify(stats));
    }

    static getStats() {
        return JSON.parse(localStorage.getItem('export_stats') || '{}');
    }

    static getMostUsed() {
        const stats = this.getStats();
        return Object.entries(stats)
            .sort(([,a], [,b]) => b.count - a.count)
            .slice(0, 5);
    }

    static getSuccessRate() {
        const stats = this.getStats();
        const total = Object.values(stats).reduce((sum, stat) => sum + stat.count, 0);
        const successful = Object.values(stats).reduce((sum, stat) => sum + stat.success, 0);

        return total > 0 ? (successful / total * 100).toFixed(1) : 0;
    }
}

// Hook per tracking automatico
const originalQuickExport = exportManager?.quickExport;
if (originalQuickExport) {
    exportManager.quickExport = async function(type, format, filters = {}) {
        try {
            const result = await originalQuickExport.call(this, type, format, filters);
            ExportAnalytics.track(type, format, true);
            return result;
        } catch (error) {
            ExportAnalytics.track(type, format, false);
            throw error;
        }
    };
}

// Console utilities per debug
window.exportDebug = {
    stats: () => ExportAnalytics.getStats(),
    mostUsed: () => ExportAnalytics.getMostUsed(),
    successRate: () => ExportAnalytics.getSuccessRate(),
    schedules: () => scheduledExport.getSchedules(),
    testExport: (type = 'hedgehogs', format = 'pdf') => quickExport(type, format),
    clearStats: () => localStorage.removeItem('export_stats'),
    clearSchedules: () => localStorage.removeItem('export_schedules')
};

console.log('🦔 La Ninna Export System loaded successfully!');
console.log('Use window.exportDebug for utilities and debugging.');

// README Documentation Finale