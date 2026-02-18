import Chat from '../components/Chat';

export default function Intelligence() {
  return (
    <div className="p-8 h-full">
      <h1 className="text-3xl font-bold text-gray-900 mb-4">NXD Intelligence</h1>
      <div className="bg-white p-8 rounded-xl shadow-sm border border-gray-200 h-full">
        <Chat />
      </div>
    </div>
  );
}
