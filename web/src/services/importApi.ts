import { api } from './api'

export interface ImportError {
  row_number: number
  field_name: string
  error_message: string
  expected_format?: string
  actual_value?: string
}

export interface ImportResult {
  success_count: number
  error_count: number
  total_rows: number
  errors?: ImportError[]
  row_errors?: Record<number, any>
}

const importApi = {
  /**
   * Generate CSV template for a CI type
   */
  async generateTemplate(ciType: string): Promise<Blob> {
    const response = await api.post('/ea/import/template', { ci_type: ciType }, {
      responseType: 'blob',
      headers: {
        Accept: 'text/csv',
      },
    })
    return response.data
  },

  /**
   * Validate import file
   */
  async validateImport(file: File, ciType: string): Promise<ImportResult> {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('ci_type', ciType)

    const response = await api.post('/ea/import/validate', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },

  /**
   * Execute import
   */
  async executeImport(file: File, ciType: string): Promise<ImportResult> {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('ci_type', ciType)

    const response = await api.post('/ea/import/execute', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },

  /**
   * Download error CSV
   */
  async downloadErrorCSV(errors: ImportError[]): Promise<Blob> {
    const response = await api.post('/ea/import/errors/download', errors, {
      responseType: 'blob',
      headers: {
        Accept: 'text/csv',
      },
    })
    return response.data
  },

  /**
   * Get import status (for async imports)
   */
  async getImportStatus(batchId: string): Promise<any> {
    const response = await api.get(`/ea/import/status/${batchId}`)
    return response.data
  },

  /**
   * Trigger browser download for a blob
   */
  downloadBlob(blob: Blob, filename: string) {
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
  },
}

export default importApi
