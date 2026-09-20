import { AppProviders } from './providers'
import { ErrorBoundary } from '../components/ErrorBoundary'
import { AppRoutes } from '../routes'

export function App() {
  return (
    <ErrorBoundary>
      <AppProviders>
        <AppRoutes />
      </AppProviders>
    </ErrorBoundary>
  )
}
