import { Disclosure} from '@headlessui/react'


let navigation = [
  { name: 'Dashboard', href: '#', current: false },
  { name: 'Team Score', href: '#', current: false },
  { name: 'Draftboard', href: '#', current: false },
  { name: 'AnotherTabBruh', href: '#', current: false },
]

function classNames(...classes) {
  return classes.filter(Boolean).join(' ')
}

export default function Example(props) {

    const logoPic = "https://www.elitefourum.com/uploads/default/original/3X/4/b/4bbe5270ed2b07d84730959af8819f255a922ea0.png";
    const backgroundColor="#2D3142"
    navigation = navigation.map(navigationPage=>{
        return{
        ...navigationPage,
        current:props.page === navigationPage.name
        }
    })

    return (
        <Disclosure style={{backgroundColor:`${backgroundColor}`}}as="nav" className="bg-gray-800">
        <div className="mx-auto max-w-7xl px-2 sm:px-6 lg:px-8">
            <div className="relative flex h-16 items-center justify-between">
            
                <div className="flex flex-1 items-center justify-center sm:items-stretch sm:justify-start">
                    <div className="flex shrink-0 items-center">
                    <img
                        alt="Logo"
                        src={logoPic}
                        className="h-8 w-auto"
                    />
                    </div>
                    <div className="hidden sm:ml-6 sm:block">
                    <div className="flex space-x-4">
                        {navigation.map((item) => (
                        <a
                            key={item.name}
                            href={item.href}
                            aria-current={item.current ? 'page' : undefined}
                            className={classNames(
                            item.current ? 'bg-gray-900 text-white' : 'text-gray-300 hover:bg-gray-700 hover:text-white',
                            'rounded-md px-3 py-2 text-sm font-medium',
                            )}
                        >
                            {item.name}
                        </a>
                        ))}
                    </div>
                    </div>
                </div>
            </div>
        </div>
        </Disclosure>
    )
}
