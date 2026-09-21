import { Route, Routes } from 'react-router-dom'
import { ActivitiesPage } from '../features/activities/pages/ActivitiesPage'
import { ActivityDetailPage } from '../features/activities/pages/ActivityDetailPage'
import { AnnouncementDetailPage } from '../features/announcements/pages/AnnouncementDetailPage'
import { AnnouncementsPage } from '../features/announcements/pages/AnnouncementsPage'
import { LoginPage } from '../features/auth/LoginPage'
import { ChatPage } from '../features/chat/pages/ChatPage'
import { ComplaintDetailPage } from '../features/complaints/pages/ComplaintDetailPage'
import { ComplaintsPage } from '../features/complaints/pages/ComplaintsPage'
import { DashboardPage } from '../features/dashboard/DashboardPage'
import { DuesPage } from '../features/dues/pages/DuesPage'
import { DuesPeriodDetailPage } from '../features/dues/pages/DuesPeriodDetailPage'
import { FinancePage } from '../features/finance/pages/FinancePage'
import { FinanceTransactionDetailPage } from '../features/finance/pages/FinanceTransactionDetailPage'
import { GalleryPage } from '../features/gallery/pages/GalleryPage'
import { LandingPage } from '../features/landing/pages/LandingPage'
import { PublicAnnouncementPage } from '../features/landing/pages/PublicAnnouncementPage'
import { ResidentsPage } from '../features/residents/pages/ResidentsPage'
import { SiteSettingsPage } from '../features/site-settings/pages/SiteSettingsPage'
import { VillagePage } from '../features/village/pages/VillagePage'
import { AdminLayout } from '../layouts/AdminLayout'
import { NotFoundPage } from './NotFoundPage'
import { ProtectedRoute } from './ProtectedRoute'

export function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/pengumuman/:id" element={<PublicAnnouncementPage />} />
      <Route path="/login" element={<LoginPage />} />

      <Route element={<ProtectedRoute />}>
        <Route element={<AdminLayout />}>
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/chat" element={<ChatPage />} />
          <Route path="/chat/:conversationId" element={<ChatPage />} />
          <Route path="/residents" element={<ResidentsPage />} />
          <Route path="/complaints" element={<ComplaintsPage />} />
          <Route path="/complaints/:id" element={<ComplaintDetailPage />} />
          <Route path="/announcements" element={<AnnouncementsPage />} />
          <Route path="/announcements/:id" element={<AnnouncementDetailPage />} />
          <Route path="/finance" element={<FinancePage />} />
          <Route path="/finance/:id" element={<FinanceTransactionDetailPage />} />
          <Route path="/dues" element={<DuesPage />} />
          <Route path="/dues/periods/:id" element={<DuesPeriodDetailPage />} />
          <Route path="/activities" element={<ActivitiesPage />} />
          <Route path="/activities/:id" element={<ActivityDetailPage />} />
          <Route path="/gallery" element={<GalleryPage />} />
          <Route path="/village" element={<VillagePage />} />
          <Route path="/site-settings" element={<SiteSettingsPage />} />
        </Route>
      </Route>

      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}
