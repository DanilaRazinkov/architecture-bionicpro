import React, { useState, useEffect } from 'react';
import { ReportsService, UserReport, UserSummary, DataAvailability } from '../services/ReportsService';

interface ReportsComponentProps {
  userId: number;
  className?: string;
}

export const ReportsComponent: React.FC<ReportsComponentProps> = ({
  userId,
  className = ''
}) => {
  const [reportsService] = useState(() => new ReportsService());
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  const [report, setReport] = useState<UserReport | null>(null);
  const [summary, setSummary] = useState<UserSummary | null>(null);
  const [dataAvailability, setDataAvailability] = useState<DataAvailability | null>(null);
  
  const [startDate, setStartDate] = useState<string>('');
  const [endDate, setEndDate] = useState<string>('');

  useEffect(() => {
    const today = new Date();
    const thirtyDaysAgo = reportsService.getDaysAgo(30);
    
    setEndDate(reportsService.formatDate(today));
    setStartDate(reportsService.formatDate(thirtyDaysAgo));
    
    loadInitialData();
  }, [userId]);

  const loadInitialData = async () => {
    setLoading(true);
    setError(null);

    try {
      const [summaryData, availabilityData] = await Promise.all([
        reportsService.getUserSummary(userId),
        reportsService.getDataAvailability()
      ]);

      setSummary(summaryData);
      setDataAvailability(availabilityData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Data loading error');
    } finally {
      setLoading(false);
    }
  };

  const handleGenerateReport = async () => {
    if (!startDate || !endDate) {
      setError('Select period for report');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const reportData = await reportsService.getUserReport(userId, startDate, endDate);
      setReport(reportData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Report generation error');
    } finally {
      setLoading(false);
    }
  };

  const getBatteryHealthColor = (health: string) => {
    switch (health) {
      case 'Excellent': return 'text-green-600';
      case 'Good': return 'text-blue-600';
      case 'Low': return 'text-yellow-600';
      case 'Critical': return 'text-red-600';
      default: return 'text-gray-600';
    }
  };

  const getUsageIntensityColor = (intensity: string) => {
    switch (intensity) {
      case 'Very High': return 'text-green-600';
      case 'High': return 'text-blue-600';
      case 'Medium': return 'text-yellow-600';
      case 'Low': return 'text-red-600';
      default: return 'text-gray-600';
    }
  };

  const getUsageIntensityBgColor = (intensity: string) => {
    switch (intensity) {
      case 'Very High': return 'bg-green-100';
      case 'High': return 'bg-blue-100';
      case 'Medium': return 'bg-yellow-100';
      case 'Low': return 'bg-red-100';
      default: return 'bg-gray-100';
    }
  };

  return (
    <div className="flex flex-col gap-6">
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-2xl font-bold text-gray-900 mb-4">
          Prosthesis Work Reports
        </h2>

        {error && (
          <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
            <div className="text-sm text-red-600">{error}</div>
          </div>
        )}

        {dataAvailability && (
          <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <h3 className="text-sm font-medium text-blue-800 mb-2">Data Availability</h3>
            <div className="text-sm text-blue-700 space-y-1">
              <p>Reports available: {dataAvailability.reports_available ? 'Yes' : 'No'}</p>
              <p>Latest report date: {dataAvailability.latest_report_date || 'No data'}</p>
              <p>Total reports: {dataAvailability.total_reports}</p>
              <p>Telemetry available: {dataAvailability.telemetry_data_available ? 'Yes' : 'No'}</p>
            </div>
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Start Date
            </label>
            <input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              max={dataAvailability?.latest_report_date || undefined}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              End Date
            </label>
            <input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              max={dataAvailability?.latest_report_date || undefined}
              className="w-full px-3 py-2 border border-gray-300 rounded-md"
            />
          </div>

          <div className="flex items-end gap-2">
            <button
              onClick={handleGenerateReport}
              disabled={loading}
              className={`flex-1 px-4 py-2 rounded-md font-medium text-white ${
                loading ? 'bg-gray-400 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-700'
              } transition-colors`}
            >
              {loading ? 'Generating...' : 'Generate'}
            </button>
          </div>
        </div>
      </div>

      <div className="bg-white shadow rounded-lg p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Summary</h3>
        {loading ? (
          <div className="text-center py-8">
            <div className="text-gray-500">Loading data...</div>
          </div>
        ) : summary && summary.total_movements !== undefined ? (
          <>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
              <div className="text-center p-4 bg-gray-50 rounded-lg">
                <div className="text-2xl font-bold text-blue-600">{summary.total_movements || 0}</div>
                <div className="text-sm text-gray-500">Total movements</div>
              </div>
              <div className="text-center p-4 bg-gray-50 rounded-lg">
                <div className="text-2xl font-bold text-green-600">{summary.usage_intensity || 'Unknown'}</div>
                <div className="text-sm text-gray-500">Usage intensity</div>
              </div>
              <div className="text-center p-4 bg-gray-50 rounded-lg">
                <div className="text-2xl font-bold text-purple-600">
                  {summary.data_quality_score ? (summary.data_quality_score * 100).toFixed(1) : '0'}%
                </div>
                <div className="text-sm text-gray-500">Data quality</div>
              </div>
            </div>
            <div className="text-sm text-gray-500 space-y-1">
              <p><strong>Username:</strong> {summary.customer_name || 'Unknown'}</p>
              <p><strong>Prosthesis type:</strong> {summary.prosthesis_type || 'Unknown'}</p>
              <p><strong>Last activity:</strong> {summary.last_activity_date || 'Unknown'}</p>
            </div>
          </>
        ) : (
          <div className="text-center py-8">
            <div className="text-gray-500">Data unavailable</div>
          </div>
        )}
      </div>

      {report && (
        <div className="bg-white shadow rounded-lg p-6">
          <div className="flex justify-between items-start mb-6">
            <div>
              <h3 className="text-lg font-medium text-gray-900">
                Report for period {report.report_period.start_date} — {report.report_period.end_date}
              </h3>
              <p className="text-sm text-gray-500 mt-1">
                Generated: {new Date(report.generated_at).toLocaleString('en-US')}
              </p>
            </div>
          </div>

          <div className="mb-6">
            <h4 className="text-base font-medium text-gray-900 mb-3">Summary Metrics</h4>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="p-3 bg-blue-50 rounded-lg">
                <div className="text-lg font-semibold text-blue-800">
                  {report.summary_metrics?.total_movements?.toLocaleString('en-US') || '0'}
                </div>
                <div className="text-xs text-blue-600">Total movements</div>
              </div>
              <div className="p-3 bg-green-50 rounded-lg">
                <div className="text-lg font-semibold text-green-800">
                  {report.summary_metrics?.avg_movement_accuracy ? (report.summary_metrics.avg_movement_accuracy * 100).toFixed(1) : '0'}%
                </div>
                <div className="text-xs text-green-600">Average accuracy</div>
              </div>
              <div className="p-3 bg-yellow-50 rounded-lg">
                <div className="text-lg font-semibold text-yellow-800">
                  {report.summary_metrics?.avg_battery_level?.toFixed(1) || '0'}%
                </div>
                <div className="text-xs text-yellow-600">Average battery level</div>
              </div>
              <div className={`p-3 ${getUsageIntensityBgColor(report.summary_usage?.usage_intensity || 'Low')} rounded-lg`}>
                <div className={`text-lg font-semibold ${getUsageIntensityColor(report.summary_usage?.usage_intensity || 'Low')}`}>
                  {report.summary_usage?.usage_intensity || 'Unknown'}
                </div>
                <div className="text-xs text-purple-600">Usage intensity</div>
              </div>
            </div>
          </div>

          <div className="mb-6">
            <h4 className="text-base font-medium text-gray-900 mb-3">Device Status</h4>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="p-3 bg-gray-50 rounded-lg">
                <div className="text-sm text-gray-500">Prosthesis type</div>
                <div className="font-medium">{report.summary_usage?.prosthesis_type || 'Unknown'}</div>
              </div>
              <div className="p-3 bg-gray-50 rounded-lg">
                <div className="text-sm text-gray-500">Primary muscle group</div>
                <div className="font-medium">{report.summary_usage?.primary_muscle_group || 'Unknown'}</div>
              </div>
              <div className="p-3 bg-gray-50 rounded-lg">
                <div className="text-sm text-gray-500">Battery status</div>
                <div className={`font-medium ${getBatteryHealthColor(report.summary_usage?.battery_health || 'Good')}`}>
                  {report.summary_usage?.battery_health || 'Unknown'}
                </div>
              </div>
            </div>
          </div>

          {report.insights && report.insights.length > 0 && (
            <div className="mb-6">
              <h4 className="text-base font-medium text-gray-900 mb-3">Analytical Insights</h4>
              <div className="flex flex-col gap-2">
                {(report.insights || []).map((insight, index) => (
                  <div key={index} className="p-3 bg-blue-50 border-l-4 border-blue-400">
                    <p className="text-sm text-blue-800">{insight}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {report.recommendations && report.recommendations.length > 0 && (
            <div className="mb-6">
              <h4 className="text-base font-medium text-gray-900 mb-3">Recommendations</h4>
              <div className="flex flex-col gap-2">
                {(report.recommendations || []).map((recommendation, index) => (
                  <div key={index} className="p-3 bg-green-50 border-l-4 border-green-400">
                    <p className="text-sm text-green-800">{recommendation}</p>
                  </div>
                ))}
              </div>
            </div>
          )}

          {report.daily_metrics && report.daily_metrics.length > 0 && (
            <div>
              <h4 className="text-base font-medium text-gray-900 mb-3">Daily Activity</h4>
              <div className="overflow-x-auto">
                <table className="min-w-full border-collapse">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Date
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Movements
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Accuracy
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Battery
                      </th>
                      <th className="px-3 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Intensity
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white">
                    {(report.daily_metrics || []).slice(-10).map((day, index) => (
                      <tr key={index} className="border-t border-gray-200">
                        <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-900">
                          {day.report_date}
                        </td>
                        <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-900">
                          {day.telemetry?.total_movements?.toLocaleString('en-US') || '0'}
                        </td>
                        <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-900">
                          {day.telemetry?.avg_movement_accuracy ? (day.telemetry.avg_movement_accuracy * 100).toFixed(1) : '0'}%
                        </td>
                        <td className="px-3 py-2 whitespace-nowrap text-sm text-gray-900">
                          {day.telemetry?.avg_battery_level?.toFixed(1) || '0'}%
                        </td>
                        <td className="px-3 py-2 whitespace-nowrap text-sm">
                          <span className={getUsageIntensityColor(day.usage?.usage_intensity || 'Low')}>
                            {day.usage?.usage_intensity || 'Unknown'}
                          </span>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
