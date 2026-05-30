import { Suspense, lazy } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ThemeProvider } from './theme/ThemeProvider'
import { ToastProvider } from './components/Toast'
import { AppLayout } from './layout/AppLayout'
import { BrowsePage } from './pages/BrowsePage'
import { AgentDetailPage } from './pages/AgentDetailPage'
import { UploadPage } from './pages/UploadPage'

const EditAgentPage = lazy(() => import('./pages/EditAgentPage'))

function PageLoader() {
  return (
    <div className="state-container loading">
      <div className="skeleton-card" style={{ width: '100%', height: '400px' }} />
    </div>
  )
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 30000,
    },
  },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>
        <ToastProvider>
          <BrowserRouter>
            <Routes>
              <Route element={<AppLayout />}>
                <Route path="/" element={<BrowsePage />} />
                <Route path="/agents/:agentName" element={<AgentDetailPage />} />
                <Route path="/agents/:agentName/edit" element={
                  <Suspense fallback={<PageLoader />}>
                    <EditAgentPage />
                  </Suspense>
                } />
                <Route path="/upload" element={<UploadPage />} />
              </Route>
            </Routes>
          </BrowserRouter>
        </ToastProvider>
      </ThemeProvider>
    </QueryClientProvider>
  )
}
