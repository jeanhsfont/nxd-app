export default function Dashboard() {
  return (
    <div className="p-8">
      <h1 className="text-3xl font-bold text-gray-900 mb-4">Dashboard Principal</h1>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-200">
          <h3 className="text-gray-500 text-sm font-medium">Status Geral</h3>
          <p className="text-2xl font-bold text-green-600 mt-2">Operacional</p>
        </div>
        {/* Placeholders */}
      </div>
    </div>
  );
}
