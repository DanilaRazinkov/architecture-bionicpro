import React, { useEffect, useState } from 'react';

type ReportData = any;

const ReportPage: React.FC = () => {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [report, setReport] = useState<ReportData | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        setLoading(true);
        const statusRes = await fetch('http://localhost:5001/api/auth/status', {
          credentials: 'include'
        });

        const status = await statusRes.json().catch(() => ({} as any));
        if (!status?.isAuthenticated) {
          setError('Not authorized');
          return;
        }

        const res = await fetch('http://localhost:5001/api/reports', {
          credentials: 'include'
        });

        if (res.status === 401) {
          setError('Not authorized');
          return;
        }

        if (res.status === 403) {
          setError('Access Restricted');
          return;
        }

        if (!res.ok) {
          if (res.status === 404) {
            setError('Report not ready yet');
            return;
          }
          setError(`Report loading error: ${res.status}`);
          return;
        }

        const data = await res.json().catch(() => null);
        setReport(data);
      } catch (e) {
        setError('Error getting report');
      } finally {
        setLoading(false);
      }
    };

    load();
  }, []);


  if (loading) {
    return (
      <div className="loading-container">
        <div className="text-center">
          <div className="loading-spinner"></div>
          <p className="loading-text">Loading...</p>
        </div>
      </div>
    );
  }

  if (error === 'Not authorized') {
    return (
      <div className="loading-container">
        <div className="text-center">
          <p className="auth-error">'Authentication error'</p>
          <div className="flex items-center justify-center gap-3">
            <button
              onClick={() => window.location.href = 'http://localhost:5001/login'}
              className="button-primary"
            >
              Login
            </button>
          </div>
        </div>
      </div>
    );
  } else if (error === 'Report not ready yet') {
     return (
       <div className="loading-container">
         <div className="text-center">
           <p className="auth-error">'Report not ready yet'</p>
           <div className="flex justify-between items-center mb-6">
           </div>
         </div>
       </div>
     );
  } else if (error === 'Access Restricted') {
     return (
       <div className="loading-container">
         <div className="text-center">
           <p className="auth-error">'Access Restricted'</p>
           <div className="flex justify-between items-center mb-6">
           </div>
         </div>
       </div>
     );
  }

  return (
    <div className="report-container">
      <div className="report-content">
        <div className="flex justify-between items-center mb-6">
          <h1 className="report-title">Your Report</h1>
        </div>
        {report ? (
          <pre className="report-pre">
            {JSON.stringify(report, null, 2)}
          </pre>
        ) : (
          <p className="report-message">Report generated and provided as file.</p>
        )}
        <div className="mt-4">
          <a href="http://localhost:5001/api/reports" className="download-link">
            Download JSON Report
          </a>
        </div>
      </div>
    </div>
  );
};

export default ReportPage;
