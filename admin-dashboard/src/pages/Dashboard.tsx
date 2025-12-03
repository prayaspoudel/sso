import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Users, Building2, Activity, Bell } from 'lucide-react';
import { apiClient } from '../lib/axios';
import { API_ENDPOINTS } from '../config/api';
import type { UserStats, CompanyStats, AuditLogStats, NotificationStats } from '../types';

const Dashboard: React.FC = () => {
  const { data: userStats } = useQuery<UserStats>({
    queryKey: ['userStats'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.USERS.STATS);
      return response.data;
    },
  });

  const { data: companyStats } = useQuery<CompanyStats>({
    queryKey: ['companyStats'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.COMPANIES.STATS);
      return response.data;
    },
  });

  const { data: auditLogStats } = useQuery<AuditLogStats>({
    queryKey: ['auditLogStats'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.AUDIT_LOGS.STATS);
      return response.data;
    },
  });

  const { data: notificationStats } = useQuery<NotificationStats>({
    queryKey: ['notificationStats'],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.NOTIFICATIONS.STATS);
      return response.data;
    },
  });

  const statsCards = [
    {
      title: 'Total Users',
      value: userStats?.total_users || 0,
      subtitle: `${userStats?.active_users || 0} active`,
      icon: Users,
      color: 'bg-blue-500',
    },
    {
      title: 'Total Companies',
      value: companyStats?.total_companies || 0,
      subtitle: `${companyStats?.active_companies || 0} active`,
      icon: Building2,
      color: 'bg-green-500',
    },
    {
      title: 'Audit Logs',
      value: auditLogStats?.total_logs || 0,
      subtitle: `${auditLogStats?.success_count || 0} successful`,
      icon: Activity,
      color: 'bg-purple-500',
    },
    {
      title: 'Notifications',
      value: notificationStats?.total_notifications || 0,
      subtitle: `${notificationStats?.unread_count || 0} unread`,
      icon: Bell,
      color: 'bg-orange-500',
    },
  ];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>

      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {statsCards.map((card) => (
          <div
            key={card.title}
            className="overflow-hidden rounded-lg bg-white shadow"
          >
            <div className="p-5">
              <div className="flex items-center">
                <div className={`flex-shrink-0 rounded-md ${card.color} p-3`}>
                  <card.icon className="h-6 w-6 text-white" />
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="truncate text-sm font-medium text-gray-500">
                      {card.title}
                    </dt>
                    <dd className="flex items-baseline">
                      <div className="text-2xl font-semibold text-gray-900">
                        {card.value.toLocaleString()}
                      </div>
                    </dd>
                    <dd className="text-xs text-gray-500">{card.subtitle}</dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div className="rounded-lg bg-white p-6 shadow">
          <h3 className="mb-4 text-lg font-semibold text-gray-900">
            Users by Role
          </h3>
          <div className="space-y-2">
            {Object.entries(userStats?.users_by_role || {}).map(([role, count]) => (
              <div key={role} className="flex justify-between">
                <span className="text-gray-600 capitalize">{role}</span>
                <span className="font-medium text-gray-900">{count}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="rounded-lg bg-white p-6 shadow">
          <h3 className="mb-4 text-lg font-semibold text-gray-900">
            Recent Activity
          </h3>
          <div className="space-y-2">
            {Object.entries(auditLogStats?.actions_by_type || {}).slice(0, 5).map(([action, count]) => (
              <div key={action} className="flex justify-between">
                <span className="text-gray-600">{action}</span>
                <span className="font-medium text-gray-900">{count}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
