import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Bell, Check, Trash2, Settings } from 'lucide-react';
import { apiClient } from '../lib/axios';
import { API_ENDPOINTS } from '../config/api';
import { useWebSocket } from '../hooks/useWebSocket';
import type { Notification, PaginatedResponse } from '../types';

const Notifications: React.FC = () => {
  const queryClient = useQueryClient();
  const { notifications: wsNotifications, markAsRead: wsMarkAsRead } = useWebSocket();
  const [page, setPage] = useState(1);
  const [filter, setFilter] = useState<'all' | 'unread' | 'read'>('all');
  const limit = 20;

  const { data, isLoading } = useQuery<PaginatedResponse<Notification>>({
    queryKey: ['notifications', page, filter],
    queryFn: async () => {
      const response = await apiClient.get(API_ENDPOINTS.NOTIFICATIONS.LIST, {
        params: { 
          page, 
          limit,
          status: filter === 'all' ? undefined : filter,
        },
      });
      return response.data;
    },
  });

  const markAsReadMutation = useMutation({
    mutationFn: async (notificationId: string) => {
      await apiClient.put(API_ENDPOINTS.NOTIFICATIONS.MARK_READ(notificationId));
    },
    onSuccess: (_, notificationId) => {
      wsMarkAsRead(notificationId);
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const markAllAsReadMutation = useMutation({
    mutationFn: async () => {
      await apiClient.put(API_ENDPOINTS.NOTIFICATIONS.MARK_ALL_READ);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const deleteNotificationMutation = useMutation({
    mutationFn: async (notificationId: string) => {
      await apiClient.delete(API_ENDPOINTS.NOTIFICATIONS.DELETE(notificationId));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications'] });
    },
  });

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'critical':
        return 'border-red-500 bg-red-50';
      case 'high':
        return 'border-orange-500 bg-orange-50';
      case 'normal':
        return 'border-blue-500 bg-blue-50';
      case 'low':
        return 'border-gray-500 bg-gray-50';
      default:
        return 'border-gray-500 bg-gray-50';
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Notifications</h1>
        <div className="flex space-x-2">
          <button
            onClick={() => markAllAsReadMutation.mutate()}
            disabled={markAllAsReadMutation.isPending}
            className="flex items-center rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <Check className="mr-2 h-4 w-4" />
            Mark All Read
          </button>
          <button className="flex items-center rounded-lg border border-gray-300 bg-white px-4 py-2 text-gray-700 hover:bg-gray-50">
            <Settings className="mr-2 h-4 w-4" />
            Preferences
          </button>
        </div>
      </div>

      <div className="flex space-x-2 rounded-lg bg-white p-2 shadow">
        <button
          onClick={() => setFilter('all')}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium ${
            filter === 'all'
              ? 'bg-blue-100 text-blue-700'
              : 'text-gray-600 hover:bg-gray-100'
          }`}
        >
          All
        </button>
        <button
          onClick={() => setFilter('unread')}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium ${
            filter === 'unread'
              ? 'bg-blue-100 text-blue-700'
              : 'text-gray-600 hover:bg-gray-100'
          }`}
        >
          Unread
        </button>
        <button
          onClick={() => setFilter('read')}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium ${
            filter === 'read'
              ? 'bg-blue-100 text-blue-700'
              : 'text-gray-600 hover:bg-gray-100'
          }`}
        >
          Read
        </button>
      </div>

      {wsNotifications.length > 0 && (
        <div className="rounded-lg bg-blue-50 p-4 shadow">
          <h3 className="mb-2 font-semibold text-blue-900">
            Real-time Notifications ({wsNotifications.length})
          </h3>
          <div className="space-y-2">
            {wsNotifications.slice(0, 3).map((notification) => (
              <div
                key={notification.id}
                className="flex items-start justify-between rounded-md bg-white p-3"
              >
                <div className="flex-1">
                  <p className="font-medium text-gray-900">{notification.title}</p>
                  <p className="text-sm text-gray-600">{notification.message}</p>
                </div>
                <button
                  onClick={() => markAsReadMutation.mutate(notification.id)}
                  className="ml-2 text-blue-600 hover:text-blue-800"
                  aria-label="Mark as read"
                >
                  <Check className="h-4 w-4" />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="space-y-3">
        {isLoading ? (
          <div className="flex h-64 items-center justify-center rounded-lg bg-white shadow">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-blue-500 border-t-transparent"></div>
          </div>
        ) : (
          <>
            {data?.data.map((notification) => (
              <div
                key={notification.id}
                className={`flex items-start justify-between rounded-lg border-l-4 p-4 shadow ${getPriorityColor(
                  notification.priority
                )} ${notification.status === 'unread' ? 'bg-white' : 'opacity-60'}`}
              >
                <div className="flex flex-1 items-start space-x-3">
                  <Bell className="mt-1 h-5 w-5 flex-shrink-0 text-gray-400" />
                  <div className="flex-1">
                    <div className="flex items-center space-x-2">
                      <h4 className="font-semibold text-gray-900">{notification.title}</h4>
                      {notification.status === 'unread' && (
                        <span className="h-2 w-2 rounded-full bg-blue-500"></span>
                      )}
                      <span className="text-xs text-gray-500 capitalize">
                        {notification.priority}
                      </span>
                    </div>
                    <p className="mt-1 text-sm text-gray-700">{notification.message}</p>
                    <p className="mt-1 text-xs text-gray-500">
                      {new Date(notification.created_at).toLocaleString()}
                    </p>
                  </div>
                </div>
                <div className="flex space-x-2">
                  {notification.status === 'unread' && (
                    <button
                      onClick={() => markAsReadMutation.mutate(notification.id)}
                      disabled={markAsReadMutation.isPending}
                      className="text-blue-600 hover:text-blue-800 disabled:cursor-not-allowed disabled:opacity-50"
                      aria-label="Mark as read"
                    >
                      <Check className="h-4 w-4" />
                    </button>
                  )}
                  <button
                    onClick={() => deleteNotificationMutation.mutate(notification.id)}
                    disabled={deleteNotificationMutation.isPending}
                    className="text-red-600 hover:text-red-800 disabled:cursor-not-allowed disabled:opacity-50"
                    aria-label="Delete notification"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            ))}

            {data && data.total_pages > 1 && (
              <div className="flex items-center justify-between rounded-lg bg-white px-4 py-3 shadow sm:px-6">
                <div className="flex flex-1 justify-between sm:hidden">
                  <button
                    onClick={() => setPage(Math.max(1, page - 1))}
                    disabled={page === 1}
                    className="relative inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    Previous
                  </button>
                  <button
                    onClick={() => setPage(Math.min(data.total_pages, page + 1))}
                    disabled={page === data.total_pages}
                    className="relative ml-3 inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    Next
                  </button>
                </div>
                <div className="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
                  <div>
                    <p className="text-sm text-gray-700">
                      Showing <span className="font-medium">{(page - 1) * limit + 1}</span> to{' '}
                      <span className="font-medium">
                        {Math.min(page * limit, data.total)}
                      </span>{' '}
                      of <span className="font-medium">{data.total}</span> results
                    </p>
                  </div>
                  <div>
                    <nav className="isolate inline-flex -space-x-px rounded-md shadow-sm">
                      <button
                        onClick={() => setPage(Math.max(1, page - 1))}
                        disabled={page === 1}
                        className="relative inline-flex items-center rounded-l-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        Previous
                      </button>
                      <button
                        onClick={() => setPage(Math.min(data.total_pages, page + 1))}
                        disabled={page === data.total_pages}
                        className="relative inline-flex items-center rounded-r-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        Next
                      </button>
                    </nav>
                  </div>
                </div>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default Notifications;
